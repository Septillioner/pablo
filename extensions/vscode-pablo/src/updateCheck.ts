import { spawn } from 'child_process';
import * as path from 'path';
import * as vscode from 'vscode';

const UPDATE_CHECK_ARGS = ['update', 'check', '--json'] as const;
const UPDATE_INSTALL_ARGS = ['update'] as const;
const UPDATE_ACTION_LABEL = 'Update';
const UPDATE_CHECK_TIMEOUT_MS = 15_000;
const UPDATE_INSTALL_TIMEOUT_MS = 120_000;

export interface CliUpdateCheckResult {
	currentVersion: string;
	latestVersion: string;
	releaseTag: string;
	updateAvailable: boolean;
}

interface UpdateCheckJson {
	current_version?: string;
	latest_version?: string;
	release_tag?: string;
	update_available?: boolean;
}

export interface CliUpdateFlowOptions {
	binaryPath: string;
	outputChannel: vscode.OutputChannel;
	onBeforeUpdate: () => Promise<void>;
	onAfterUpdate: () => Promise<void>;
}

function runPablo(
	binaryPath: string,
	args: readonly string[],
	timeoutMs: number
): Promise<{ stdout: string; code: number }> {
	const resolved = path.normalize(binaryPath);

	return new Promise((resolve) => {
		const child = spawn(resolved, [...args], { shell: false, windowsHide: true });
		let stdout = '';
		child.stdout?.on('data', (chunk: Buffer) => {
			stdout += chunk.toString();
		});
		child.stderr?.on('data', (chunk: Buffer) => {
			stdout += chunk.toString();
		});
		const timer = setTimeout(() => {
			child.kill();
			resolve({ stdout, code: -1 });
		}, timeoutMs);
		child.on('close', (code) => {
			clearTimeout(timer);
			resolve({ stdout, code: code ?? -1 });
		});
		child.on('error', () => {
			clearTimeout(timer);
			resolve({ stdout, code: -1 });
		});
	});
}

function parseUpdateCheckJson(raw: string): CliUpdateCheckResult | undefined {
	const trimmed = raw.trim();
	if (!trimmed) {
		return undefined;
	}

	const jsonStart = trimmed.indexOf('{');
	const jsonEnd = trimmed.lastIndexOf('}');
	if (jsonStart < 0 || jsonEnd <= jsonStart) {
		return undefined;
	}

	try {
		const parsed = JSON.parse(trimmed.slice(jsonStart, jsonEnd + 1)) as UpdateCheckJson;
		if (typeof parsed.update_available !== 'boolean') {
			return undefined;
		}
		return {
			currentVersion: parsed.current_version ?? '',
			latestVersion: parsed.latest_version ?? '',
			releaseTag: parsed.release_tag ?? '',
			updateAvailable: parsed.update_available,
		};
	} catch {
		return undefined;
	}
}

export async function checkForCliUpdate(binaryPath: string): Promise<CliUpdateCheckResult | undefined> {
	const { stdout } = await runPablo(binaryPath, UPDATE_CHECK_ARGS, UPDATE_CHECK_TIMEOUT_MS);
	return parseUpdateCheckJson(stdout);
}

export async function installCliUpdate(binaryPath: string): Promise<{ ok: boolean; output: string }> {
	const { stdout, code } = await runPablo(binaryPath, UPDATE_INSTALL_ARGS, UPDATE_INSTALL_TIMEOUT_MS);
	return { ok: code === 0, output: stdout };
}

/** Activation path: silent when up to date; toast only when an update is available. */
export async function notifyCliUpdateIfAvailable(options: CliUpdateFlowOptions): Promise<void> {
	await runCliUpdateFlow({ ...options, interactive: false });
}

/** Manual "Pablo: Update" command: always report status to the user. */
export async function runCliUpdateCommand(options: CliUpdateFlowOptions): Promise<void> {
	await runCliUpdateFlow({ ...options, interactive: true });
}

async function runCliUpdateFlow(
	options: CliUpdateFlowOptions & { interactive: boolean }
): Promise<void> {
	const { binaryPath, outputChannel, onBeforeUpdate, onAfterUpdate, interactive } = options;

	let result: CliUpdateCheckResult | undefined;
	try {
		result = await checkForCliUpdate(binaryPath);
	} catch (err: unknown) {
		const message = err instanceof Error ? err.message : String(err);
		outputChannel.appendLine(`CLI update check failed: ${message}`);
		if (interactive) {
			vscode.window.showWarningMessage(
				`Could not check for Pablo CLI updates: ${message}`
			);
		}
		return;
	}

	if (!result) {
		const skipMessage =
			'CLI update check skipped (offline, unsupported CLI, or parse error).';
		outputChannel.appendLine(skipMessage);
		if (interactive) {
			vscode.window.showWarningMessage(
				'Could not check for Pablo CLI updates. You may be offline, or this CLI does not support update check.'
			);
		}
		return;
	}

	if (!result.updateAvailable) {
		const upToDate = `Pablo CLI is up to date (${result.currentVersion}).`;
		outputChannel.appendLine(upToDate);
		if (interactive) {
			vscode.window.showInformationMessage(upToDate);
		}
		return;
	}

	const summary = `Pablo CLI update available: ${result.currentVersion} -> ${result.latestVersion}`;
	outputChannel.appendLine(summary);

	const choice = await vscode.window.showInformationMessage(summary, UPDATE_ACTION_LABEL);
	if (choice !== UPDATE_ACTION_LABEL) {
		return;
	}

	await vscode.window.withProgress(
		{
			location: vscode.ProgressLocation.Notification,
			title: 'Updating Pablo CLI...',
			cancellable: false,
		},
		async () => {
			try {
				await onBeforeUpdate();
				const install = await installCliUpdate(binaryPath);
				if (!install.ok) {
					outputChannel.appendLine(`CLI update failed:\n${install.output}`);
					vscode.window.showErrorMessage(
						'Pablo CLI update failed. Check Output > Pablo Language Server for details.'
					);
					return;
				}
				outputChannel.appendLine(`CLI update succeeded:\n${install.output}`);
				vscode.window.showInformationMessage(
					`Pablo CLI updated to ${result.latestVersion}.`
				);
			} finally {
				await onAfterUpdate();
			}
		}
	);
}
