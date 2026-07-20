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

export async function notifyCliUpdateIfAvailable(options: {
	binaryPath: string;
	outputChannel: vscode.OutputChannel;
	onBeforeUpdate: () => Promise<void>;
	onAfterUpdate: () => Promise<void>;
}): Promise<void> {
	const { binaryPath, outputChannel, onBeforeUpdate, onAfterUpdate } = options;

	let result: CliUpdateCheckResult | undefined;
	try {
		result = await checkForCliUpdate(binaryPath);
	} catch (err: unknown) {
		const message = err instanceof Error ? err.message : String(err);
		outputChannel.appendLine(`CLI update check failed: ${message}`);
		return;
	}

	if (!result) {
		outputChannel.appendLine('CLI update check skipped (offline, unsupported CLI, or parse error).');
		return;
	}

	if (!result.updateAvailable) {
		outputChannel.appendLine(`Pablo CLI is up to date (${result.currentVersion}).`);
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
