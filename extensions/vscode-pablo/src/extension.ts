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

function resolvePabloBinary(context: vscode.ExtensionContext, outputChannel: vscode.OutputChannel): string | undefined {
	const configured = vscode.workspace.getConfiguration('pablo').get<string>('path');
	if (configured && configured.trim() !== '') {
		const configuredPath = configured.trim();
		if (fs.existsSync(configuredPath)) {
			return configuredPath;
		}
		outputChannel.appendLine(`Configured pablo.path not found: ${configuredPath}`);
	}

	const platform = process.platform;
	const arch = process.arch;
	const bundledName = `pablo-${platform}-${arch}${platform === 'win32' ? '.exe' : ''}`;
	const bundledCandidates = [
		path.resolve(context.extensionPath, 'bin', bundledName),
		path.resolve(context.extensionPath, 'bin', platform === 'win32' ? 'pablo.exe' : 'pablo'),
	];

	for (const candidate of bundledCandidates) {
		if (fs.existsSync(candidate)) {
			return candidate;
		}
	}

	return 'pablo';
}

export function activate(context: vscode.ExtensionContext) {
	const outputChannel = vscode.window.createOutputChannel('Pablo Language Server');
	outputChannel.appendLine('Pablo extension is now active!');

	const pabloBinary = resolvePabloBinary(context, outputChannel);
	if (!pabloBinary) {
		vscode.window.showErrorMessage('Pablo binary not found. Set pablo.path or install pablo on PATH.');
		return;
	}

	outputChannel.appendLine(`Using Pablo binary: ${pabloBinary}`);

	const serverOptions: ServerOptions = {
		run: { command: pabloBinary, args: ['lsp'], transport: TransportKind.stdio },
		debug: { command: pabloBinary, args: ['lsp'], transport: TransportKind.stdio }
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
		args = ` -f "${editor.document.fileName}"`;
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
