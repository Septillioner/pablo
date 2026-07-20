import * as path from 'path';
import * as vscode from 'vscode';

export type ShellKind = 'powershell' | 'cmd' | 'posix';

export function detectShellKind(): ShellKind {
	const shellPath = vscode.env.shell;
	if (!shellPath) {
		return process.platform === 'win32' ? 'powershell' : 'posix';
	}

	const base = path.basename(shellPath).toLowerCase();
	if (base.includes('powershell') || base === 'pwsh.exe' || base === 'pwsh') {
		return 'powershell';
	}
	if (base === 'cmd.exe' || base === 'cmd') {
		return 'cmd';
	}
	return 'posix';
}

const PABLO_SUBCOMMANDS = new Set([
	'run',
	'check',
	'init',
	'inspect',
	'uninstall',
	'version',
	'lsp',
	'update',
]);

export function quoteForShell(value: string, kind: ShellKind = detectShellKind()): string {
	switch (kind) {
		case 'powershell':
			return `'${value.replace(/'/g, "''")}'`;
		case 'cmd':
			return `"${value.replace(/"/g, '\\"')}"`;
		case 'posix':
			return `'${value.replace(/'/g, "'\\''")}'`;
	}
}

function shouldQuoteArg(arg: string): boolean {
	if (arg.startsWith('-')) {
		return false;
	}
	if (PABLO_SUBCOMMANDS.has(arg)) {
		return false;
	}
	return true;
}

export function formatArgForShell(arg: string, kind: ShellKind = detectShellKind()): string {
	return shouldQuoteArg(arg) ? quoteForShell(arg, kind) : arg;
}

export function buildTerminalCommand(
	executable: string,
	args: string[],
	kind: ShellKind = detectShellKind()
): string {
	const quotedExecutable = quoteForShell(executable, kind);
	const formattedArgs = args.map((arg) => formatArgForShell(arg, kind)).join(' ');

	if (kind === 'powershell') {
		return `& ${quotedExecutable} ${formattedArgs}`;
	}
	return `${quotedExecutable} ${formattedArgs}`;
}

export function executableOpenDialogOptions(): Pick<vscode.OpenDialogOptions, 'openLabel' | 'filters'> {
	return {
		openLabel: 'Select Pablo executable',
		filters: process.platform === 'win32'
			? { Executables: ['exe'], 'All Files': ['*'] }
			: undefined,
	};
}
