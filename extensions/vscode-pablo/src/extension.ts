import * as vscode from 'vscode';
import * as path from 'path';
import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
	TransportKind
} from 'vscode-languageclient/node';

import * as fs from 'fs';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
	// Create output channel for logging
	const outputChannel = vscode.window.createOutputChannel('Pablo Language Server');
	outputChannel.show(true);
	outputChannel.appendLine('Pablo extension is now active!');

	// LSP Server configuration - use absolute path resolution
	// In production, we'll have binaries for each platform: pablo-lsp-darwin-arm64, pablo-lsp-win32-x64.exe, etc.
	const platform = process.platform; // win32, darwin, linux
	const arch = process.arch; // x64, arm64
	const binaryName = `pablo-lsp-${platform}-${arch}${platform === 'win32' ? '.exe' : ''}`;

	let serverModule = path.resolve(context.extensionPath, 'bin', binaryName);

	// Fallback to the default name for development
	if (!fs.existsSync(serverModule)) {
		const devBinary = path.resolve(context.extensionPath, 'bin', `pablo-lsp${platform === 'win32' ? '.exe' : ''}`);
		if (fs.existsSync(devBinary)) {
			serverModule = devBinary;
		}
	}

	if (fs.existsSync(serverModule)) {
		try {
			serverModule = fs.realpathSync(serverModule);
			outputChannel.appendLine(`Resolved server module path: ${serverModule}`);
		} catch (err) {
			outputChannel.appendLine(`Error resolving realpath: ${err}`);
		}
	} else {
		outputChannel.appendLine(`CRITICAL: Server binary not found at ${serverModule}`);
		vscode.window.showErrorMessage(`Pablo LSP binary not found for your platform (${platform}-${arch}).`);
		return;
	}

	const serverOptions: ServerOptions = {
		run: { command: serverModule, transport: TransportKind.stdio },
		debug: { command: serverModule, transport: TransportKind.stdio }
	};

	const clientOptions: LanguageClientOptions = {
		documentSelector: [
			{ scheme: 'file', language: 'pablo' },
			{ scheme: 'file', language: 'yaml', pattern: '**/pablo*.{yaml,yml}' }
		],
		synchronize: {
			fileEvents: vscode.workspace.createFileSystemWatcher('**/pablo*.{yaml,yml}')
		}
	};

	client = new LanguageClient(
		'pabloLSP',
		'Pablo Language Server',
		serverOptions,
		clientOptions
	);

	client.start();

	// Commands
	context.subscriptions.push(vscode.commands.registerCommand('pablo.check', () => {
		runPabloCommand('check', true);
	}));

	context.subscriptions.push(vscode.commands.registerCommand('pablo.init', () => {
		const terminal = vscode.window.terminals.find(t => t.name === 'Pablo CLI') || vscode.window.createTerminal('Pablo CLI');
		terminal.show();
		terminal.sendText('pablo init');
	}));

	context.subscriptions.push(vscode.commands.registerCommand('pablo.run', () => {
		runPabloCommand('run');
	}));
}

function runPabloCommand(command: string, fileSpecific: boolean = false) {
	const editor = vscode.window.activeTextEditor;
	let args = '';

	if (fileSpecific) {
		if (editor) {
			args = ` -f "${editor.document.fileName}"`;
		} else {
			vscode.window.showErrorMessage('No active YAML editor found.');
			return;
		}
	} else if (editor && editor.document.fileName.match(/pablo.*\.ya?ml/)) {
		args = ` -c "${editor.document.fileName}"`;
	}

	const terminal = vscode.window.terminals.find(t => t.name === 'Pablo CLI') || vscode.window.createTerminal('Pablo CLI');
	terminal.show();
	terminal.sendText(`pablo ${command}${args}`);
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
