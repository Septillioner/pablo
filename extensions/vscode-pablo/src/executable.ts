import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import { execFile, spawn } from 'child_process';
import { promisify } from 'util';
import { executableOpenDialogOptions } from './shell';

const execFileAsync = promisify(execFile);
const WORKSPACE_STATE_KEY = 'pablo.executable';
const PROBE_TIMEOUT_MS = 5000;
const WHERE_EXE = process.platform === 'win32'
	? path.join(process.env.SystemRoot ?? 'C:\\Windows', 'System32', 'where.exe')
	: 'which';

function sourceSortOrder(source: string): number {
	switch (source) {
		case 'PATH':
			return 0;
		case 'selected':
			return 1;
		case 'setting':
			return 2;
		case 'workspace':
			return 3;
		default:
			return 4;
	}
}

function sortExecutablesForPicker(
	executables: string[],
	context: vscode.ExtensionContext,
	workspaceFolder?: string
): string[] {
	return [...executables].sort((left, right) => {
		const leftOrder = sourceSortOrder(resolveCandidateSource(left, context, workspaceFolder));
		const rightOrder = sourceSortOrder(resolveCandidateSource(right, context, workspaceFolder));
		if (leftOrder !== rightOrder) {
			return leftOrder - rightOrder;
		}
		return left.localeCompare(right);
	});
}

function isExecutableCandidate(filePath: string): boolean {
	try {
		if (!fs.existsSync(filePath)) {
			return false;
		}
		const stat = fs.statSync(filePath);
		if (!stat.isFile()) {
			return false;
		}
		if (process.platform === 'win32') {
			return filePath.toLowerCase().endsWith('.exe');
		}
		return true;
	} catch {
		return false;
	}
}

export const PABLO_RELEASES_URL = 'https://github.com/septillioner/pablo/releases';

export type PabloMissingAction = 'cancel' | 'select' | 'install';

export interface PabloExecutableInfo {
	path: string;
	version?: string;
	hasLsp: boolean;
	source: string;
}

function resolveSettingPath(rawPath: string): string {
	const folders = vscode.workspace.workspaceFolders;
	const workspaceFolder = folders?.[0]?.uri.fsPath ?? '';
	return rawPath
		.replace(/\$\{workspaceFolder\}/g, workspaceFolder)
		.replace(/\$\{workspaceRoot\}/g, workspaceFolder)
		.trim();
}

function normalizePath(candidate: string): string {
	return path.normalize(candidate);
}

function addCandidate(seen: Set<string>, list: string[], candidate: string): void {
	if (!candidate) {
		return;
	}
	const normalized = normalizePath(candidate);
	if (seen.has(normalized.toLowerCase())) {
		return;
	}
	seen.add(normalized.toLowerCase());
	list.push(normalized);
}

export async function listPathCandidates(): Promise<string[]> {
	return findOnPath();
}

function workspacePickerCandidates(workspaceFolder: string): string[] {
	const candidates: string[] = [];
	const buildDir = path.join(workspaceFolder, 'build');
	const defaultName = process.platform === 'win32' ? 'pablo.exe' : 'pablo';
	const defaultBuild = path.join(buildDir, defaultName);
	if (isExecutableCandidate(defaultBuild)) {
		candidates.push(defaultBuild);
	}

	if (fs.existsSync(buildDir)) {
		for (const name of fs.readdirSync(buildDir)) {
			if (!name.startsWith('pablo')) {
				continue;
			}
			const fullPath = path.join(buildDir, name);
			if (isExecutableCandidate(fullPath)) {
				candidates.push(fullPath);
			}
		}
	}

	const releasesDir = path.join(workspaceFolder, 'dist', 'releases');
	if (fs.existsSync(releasesDir)) {
		for (const name of fs.readdirSync(releasesDir)) {
			if (name.startsWith('pablo')) {
				const fullPath = path.join(releasesDir, name);
				const stat = fs.statSync(fullPath);
				if (stat.isFile()) {
					candidates.push(fullPath);
				}
			}
		}
	}

	return candidates;
}

async function findOnPath(): Promise<string[]> {
	try {
		if (process.platform === 'win32') {
			const { stdout } = await execFileAsync(WHERE_EXE, ['pablo'], {
				timeout: PROBE_TIMEOUT_MS,
				windowsHide: true
			});
			const results = stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
			return results.filter(isExecutableCandidate);
		}
		const { stdout } = await execFileAsync('which', ['-a', 'pablo'], { timeout: PROBE_TIMEOUT_MS });
		return stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
	} catch {
		return [];
	}
}

function discoverPickerCandidates(context: vscode.ExtensionContext): string[] {
	const seen = new Set<string>();
	const candidates: string[] = [];
	const workspaceFolder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;

	const stored = getSelectedExecutable(context);
	if (stored && fs.existsSync(stored)) {
		addCandidate(seen, candidates, stored);
	}

	const configured = vscode.workspace.getConfiguration('pablo').get<string>('path');
	if (configured?.trim()) {
		const configuredPath = resolveSettingPath(configured);
		if (fs.existsSync(configuredPath)) {
			addCandidate(seen, candidates, configuredPath);
		}
	}

	if (workspaceFolder) {
		for (const candidate of workspacePickerCandidates(workspaceFolder)) {
			if (fs.existsSync(candidate)) {
				addCandidate(seen, candidates, candidate);
			}
		}
	}

	return candidates;
}

function runCommand(binary: string, args: string[]): Promise<{ stdout: string; code: number; spawnError?: string }> {
	const resolved = path.normalize(binary);
	if (!isExecutableCandidate(resolved)) {
		return Promise.resolve({ stdout: '', code: -1, spawnError: 'not-executable' });
	}

	return new Promise((resolve) => {
		const child = spawn(resolved, args, { shell: false, windowsHide: true });
		let stdout = '';
		child.stdout?.on('data', (chunk: Buffer) => {
			stdout += chunk.toString();
		});
		child.stderr?.on('data', (chunk: Buffer) => {
			stdout += chunk.toString();
		});
		const timer = setTimeout(() => {
			child.kill();
			resolve({ stdout, code: -1, spawnError: 'timeout' });
		}, PROBE_TIMEOUT_MS);
		child.on('close', (code) => {
			clearTimeout(timer);
			resolve({ stdout, code: code ?? -1 });
		});
		child.on('error', (err) => {
			clearTimeout(timer);
			resolve({ stdout, code: -1, spawnError: err.message });
		});
	});
}

export async function probePabloVersion(binaryPath: string): Promise<string | undefined> {
	const { stdout, code } = await runCommand(binaryPath, ['version']);
	if (code !== 0) {
		return undefined;
	}
	const match = stdout.match(/Pablo Version:\s*(\S+)/i) ?? stdout.match(/\bv(\d+\.\d+\.\d+)\b/i);
	return match?.[1];
}

export async function assertLspSupported(binaryPath: string): Promise<boolean> {
	const { stdout, code } = await runCommand(binaryPath, ['help', 'lsp']);
	if (code !== 0) {
		return false;
	}
	if (/unknown (command|help topic)/i.test(stdout)) {
		return false;
	}
	return /language server|stdio/i.test(stdout);
}

export function getSelectedExecutable(context: vscode.ExtensionContext): string | undefined {
	return context.workspaceState.get<string>(WORKSPACE_STATE_KEY);
}

export async function setSelectedExecutable(context: vscode.ExtensionContext, executablePath: string): Promise<void> {
	await context.workspaceState.update(WORKSPACE_STATE_KEY, normalizePath(executablePath));
}

export async function clearSelectedExecutable(context: vscode.ExtensionContext): Promise<void> {
	await context.workspaceState.update(WORKSPACE_STATE_KEY, undefined);
}

export function normalizeExecutablePath(binary: string): string {
	const normalized = normalizePath(binary);
	if (path.isAbsolute(normalized) && fs.existsSync(normalized)) {
		return normalized;
	}

	const workspaceFolder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
	if (workspaceFolder) {
		const fromWorkspace = path.resolve(workspaceFolder, normalized);
		if (fs.existsSync(fromWorkspace)) {
			return normalizePath(fromWorkspace);
		}
	}

	const resolved = path.resolve(normalized);
	if (fs.existsSync(resolved)) {
		return normalizePath(resolved);
	}

	return normalized;
}

export async function resolvePabloBinary(
	context: vscode.ExtensionContext,
	outputChannel?: vscode.OutputChannel
): Promise<string | undefined> {
	const stored = getSelectedExecutable(context);
	if (stored && fs.existsSync(stored)) {
		return normalizeExecutablePath(stored);
	}
	if (stored) {
		outputChannel?.appendLine(`Stored Pablo executable not found: ${stored}`);
	}

	const configured = vscode.workspace.getConfiguration('pablo').get<string>('path');
	if (configured?.trim()) {
		const configuredPath = resolveSettingPath(configured);
		if (fs.existsSync(configuredPath)) {
			return normalizeExecutablePath(configuredPath);
		}
		outputChannel?.appendLine(`Configured pablo.path not found: ${configuredPath}`);
	}

	const onPath = await findOnPath();
	if (onPath.length === 1) {
		return normalizeExecutablePath(onPath[0]);
	}
	if (onPath.length > 1) {
		outputChannel?.appendLine(`Multiple Pablo binaries on PATH (${onPath.length}). Use Pablo: Select Executable.`);
		return undefined;
	}

	return undefined;
}

export async function promptPabloMissing(): Promise<PabloMissingAction> {
	const choice = await vscode.window.showWarningMessage(
		'Pablo CLI not found or does not support LSP (pablo lsp). Pablo 1.3+ is required.',
		{ modal: true },
		'Cancel',
		'Select Pablo',
		'Install Pablo'
	);

	if (choice === 'Select Pablo') {
		return 'select';
	}
	if (choice === 'Install Pablo') {
		return 'install';
	}
	return 'cancel';
}

function formatPickLabel(info: PabloExecutableInfo): string {
	const version = info.version ? ` ${info.version}` : '';
	const lsp = info.hasLsp ? '' : ' (no lsp)';
	return `${info.path}${version}${lsp}`;
}

function formatPickDescription(info: PabloExecutableInfo): string {
	return info.source;
}

interface PabloPickItem extends vscode.QuickPickItem {
	binaryPath: string;
}

function resolveCandidateSource(
	candidate: string,
	context: vscode.ExtensionContext,
	workspaceFolder?: string
): string {
	const current = getSelectedExecutable(context);
	if (candidate === current) {
		return 'selected';
	}
	if (workspaceFolder && candidate.startsWith(normalizePath(workspaceFolder))) {
		return 'workspace';
	}
	const configured = vscode.workspace.getConfiguration('pablo').get<string>('path');
	if (configured?.trim() && candidate === normalizeExecutablePath(resolveSettingPath(configured))) {
		return 'setting';
	}
	return 'PATH';
}

async function enrichQuickPickItems(
	quickPick: vscode.QuickPick<PabloPickItem>,
	executables: string[],
	context: vscode.ExtensionContext,
	workspaceFolder?: string
): Promise<void> {
	for (const candidate of executables) {
		const source = resolveCandidateSource(candidate, context, workspaceFolder);
		const version = await probePabloVersion(candidate);
		const hasLsp = await assertLspSupported(candidate);
		const info: PabloExecutableInfo = {
			path: normalizePath(candidate),
			version,
			hasLsp,
			source
		};
		const index = quickPick.items.findIndex((item) => item.binaryPath === candidate);
		if (index >= 0) {
			const next = [...quickPick.items];
			next[index] = {
				...next[index],
				label: formatPickLabel(info),
				description: formatPickDescription(info)
			};
			quickPick.items = next;
		}
	}
}

export async function showExecutablePicker(context: vscode.ExtensionContext): Promise<string | undefined> {
	try {
		const seen = new Set<string>();
		const rawCandidates = discoverPickerCandidates(context);

		for (const candidate of await findOnPath()) {
			addCandidate(seen, rawCandidates, candidate);
		}

		const workspaceFolder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
		const executables = sortExecutablesForPicker(
			rawCandidates.filter(isExecutableCandidate),
			context,
			workspaceFolder
		);
		const current = getSelectedExecutable(context);

		const quickPick = vscode.window.createQuickPick<PabloPickItem>();
		quickPick.title = 'Pablo: Select Executable';
		quickPick.placeholder = 'Select Pablo executable for LSP and CLI commands';
		quickPick.matchOnDescription = true;
		quickPick.busy = executables.length > 0;

		const items: PabloPickItem[] = executables.map((candidate) => ({
			label: candidate,
			description: resolveCandidateSource(candidate, context, workspaceFolder),
			detail: candidate === current ? 'Currently selected' : undefined,
			picked: candidate === current,
			binaryPath: candidate
		}));
		items.push({
			label: 'Browse for executable...',
			description: 'Select pablo binary from disk',
			binaryPath: ''
		});
		quickPick.items = items;
		quickPick.show();

		if (executables.length > 0) {
			void enrichQuickPickItems(quickPick, executables, context, workspaceFolder).finally(() => {
				quickPick.busy = false;
			});
		} else {
			quickPick.busy = false;
		}

		const pick = await new Promise<PabloPickItem | undefined>((resolve) => {
			quickPick.onDidAccept(() => {
				const selected = quickPick.selectedItems[0];
				quickPick.hide();
				resolve(selected);
			});
			quickPick.onDidHide(() => resolve(undefined));
		});

		if (!pick) {
			return undefined;
		}

		if (pick.label === 'Browse for executable...') {
			const uris = await vscode.window.showOpenDialog({
				canSelectMany: false,
				...executableOpenDialogOptions(),
			});
			if (!uris?.[0]) {
				return undefined;
			}
			return uris[0].fsPath;
		}

		return pick.binaryPath;
	} catch (err: unknown) {
		const message = err instanceof Error ? err.message : String(err);
		vscode.window.showErrorMessage(`Pablo: Select Executable failed (${message})`);
		return undefined;
	}
}
