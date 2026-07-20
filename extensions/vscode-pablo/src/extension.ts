import * as vscode from 'vscode';
import {
	CloseAction,
	ErrorAction,
	Executable,
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
	State,
	Trace
} from 'vscode-languageclient/node';

import {
	assertLspSupported,
	clearSelectedExecutable,
	getSelectedExecutable,
	normalizeExecutablePath,
	PABLO_RELEASES_URL,
	promptPabloMissing,
	resolvePabloBinary,
	setSelectedExecutable,
	showExecutablePicker
} from './executable';
import { buildTerminalCommand } from './shell';
import { inspectManifest } from './inspect';
import { registerProfileDecorations } from './profileDecorations';
import { notifyCliUpdateIfAvailable } from './updateCheck';

let client: LanguageClient | undefined;
let extensionContext: vscode.ExtensionContext;
let outputChannel: vscode.OutputChannel;
let traceOutputChannel: vscode.OutputChannel;
let activeBinary: string | undefined;
let isRestarting = false;

function parseTraceLevel(value: string): Trace {
	switch (value) {
		case 'verbose':
			return Trace.Verbose;
		case 'messages':
			return Trace.Messages;
		default:
			return Trace.Off;
	}
}

function createLanguageClient(pabloBinary: string): LanguageClient {
	const executable: Executable = {
		command: pabloBinary,
		args: ['lsp'],
		options: { shell: false }
	};

	const serverOptions: ServerOptions = {
		run: executable,
		debug: { ...executable }
	};

	const clientOptions: LanguageClientOptions = {
		documentSelector: [
			{ scheme: 'file', language: 'pablo' },
			{ scheme: 'file', language: 'yaml', pattern: '**/pablo*.{yaml,yml}' }
		],
		synchronize: {
			fileEvents: vscode.workspace.createFileSystemWatcher('**/pablo*.{yaml,yml}')
		},
		outputChannel,
		traceOutputChannel,
		errorHandler: {
			error: () => ({ action: ErrorAction.Shutdown, handled: true }),
			closed: () => ({ action: CloseAction.DoNotRestart, handled: true })
		},
		connectionOptions: {
			maxRestartCount: 0
		},
		middleware: {
			provideCodeLenses: async (document, token, next) => {
				try {
					const lenses = await next(document, token);
					return lenses?.map((lens) => {
						if (lens.command?.command !== 'pablo.runWithArgs') {
							return lens;
						}

						const runTarget = String(lens.command.arguments?.[1] ?? '');
						return new vscode.CodeLens(lens.range, {
							title: '$(play) Run',
							command: lens.command.command,
							arguments: lens.command.arguments,
							tooltip: runTarget,
						});
					}) ?? [];
				} catch (err: unknown) {
					const message = err instanceof Error ? err.message : String(err);
					outputChannel.appendLine(`CodeLens request failed: ${message}`);
					return [];
				}
			},
		},
	};

	return new LanguageClient(
		'pabloLSP',
		'Pablo Language Server',
		serverOptions,
		clientOptions
	);
}

function disposeLanguageClientReference(): void {
	client = undefined;
	activeBinary = undefined;
}

async function stopLanguageClientIfRunning(): Promise<void> {
	if (!client) {
		return;
	}
	if (client.state === State.Running) {
		await client.stop();
	}
	disposeLanguageClientReference();
}

async function startLanguageClient(pabloBinary: string): Promise<void> {
	const resolved = normalizeExecutablePath(pabloBinary);
	outputChannel.appendLine(`Using Pablo binary: ${resolved}`);

	if (!(await assertLspSupported(resolved))) {
		outputChannel.appendLine(`Pablo binary does not support LSP: ${resolved}`);
		throw new Error('Pablo binary does not support LSP (1.3+ required)');
	}

	client = createLanguageClient(resolved);
	activeBinary = resolved;

	const traceSetting = vscode.workspace.getConfiguration('pablo').get<string>('trace.server', 'off');
	client.setTrace(parseTraceLevel(traceSetting));
	if (traceSetting !== 'off') {
		traceOutputChannel.appendLine(`LSP trace level: ${traceSetting}`);
	}

	client.onDidChangeState((event) => {
		outputChannel.appendLine(`LSP state: ${State[event.oldState]} -> ${State[event.newState]}`);
		if (event.newState === State.Stopped && !isRestarting) {
			vscode.window.showErrorMessage(
				'Pablo Language Server stopped unexpectedly. Check Output > Pablo Language Server and Pablo LSP Trace.'
			);
		}
	});

	try {
		await client.start();
	} catch (err: unknown) {
		const message = err instanceof Error ? err.message : String(err);
		outputChannel.appendLine(`LSP start failed: ${message}`);
		if (client?.state !== State.Running) {
			disposeLanguageClientReference();
		}
		throw err;
	}
}

async function restartLanguageClient(pabloBinary: string): Promise<void> {
	const resolved = normalizeExecutablePath(pabloBinary);
	if (activeBinary === resolved && client?.state === State.Running) {
		outputChannel.appendLine(`Pablo binary unchanged: ${resolved}`);
		return;
	}

	isRestarting = true;
	try {
		await stopLanguageClientIfRunning();
		await startLanguageClient(resolved);
	} finally {
		isRestarting = false;
	}
}

async function handlePabloMissing(context: vscode.ExtensionContext): Promise<void> {
	const action = await promptPabloMissing();

	if (action === 'install') {
		await vscode.env.openExternal(vscode.Uri.parse(PABLO_RELEASES_URL));
		return;
	}

	if (action === 'select') {
		const selected = await showExecutablePicker(context);
		if (!selected) {
			return;
		}
		const resolved = normalizeExecutablePath(selected);
		await setSelectedExecutable(context, resolved);

		if (await assertLspSupported(resolved)) {
			try {
				await restartLanguageClient(resolved);
				vscode.window.showInformationMessage(`Pablo executable: ${resolved}`);
			} catch {
				await handlePabloMissing(context);
			}
		} else {
			vscode.window.showWarningMessage(
				`Selected binary does not support LSP (1.3+ required): ${resolved}`
			);
		}
	}
}

async function tryResolveAndStart(context: vscode.ExtensionContext): Promise<boolean> {
	const pabloBinary = await resolvePabloBinary(context, outputChannel);
	if (!pabloBinary) {
		outputChannel.appendLine('Pablo binary not found or ambiguous.');
		return false;
	}

	if (await assertLspSupported(pabloBinary)) {
		try {
			await startLanguageClient(pabloBinary);
			return true;
		} catch {
			return false;
		}
	}

	outputChannel.appendLine(`Pablo binary does not support LSP: ${pabloBinary}`);

	const stored = getSelectedExecutable(context);
	if (stored && normalizeExecutablePath(stored) === normalizeExecutablePath(pabloBinary)) {
		await clearSelectedExecutable(context);
		outputChannel.appendLine('Cleared invalid stored Pablo executable.');
		return tryResolveAndStart(context);
	}

	return false;
}

async function ensureLanguageClientStarted(context: vscode.ExtensionContext): Promise<void> {
	const started = await tryResolveAndStart(context);
	if (!started) {
		await handlePabloMissing(context);
	}
}

export async function activate(context: vscode.ExtensionContext) {
	extensionContext = context;
	outputChannel = vscode.window.createOutputChannel('Pablo Language Server');
	traceOutputChannel = vscode.window.createOutputChannel('Pablo LSP Trace');
	outputChannel.appendLine('Pablo extension is now active!');
	registerProfileDecorations(context);

	context.subscriptions.push(
		vscode.commands.registerCommand('pablo.selectExecutable', async () => {
			outputChannel.appendLine('Opening Pablo executable picker...');
			const selected = await showExecutablePicker(context);
			if (!selected) {
				return;
			}
			const resolved = normalizeExecutablePath(selected);
			await setSelectedExecutable(context, resolved);

			if (await assertLspSupported(resolved)) {
				try {
					await restartLanguageClient(resolved);
					vscode.window.showInformationMessage(`Pablo executable: ${resolved}`);
				} catch {
					await handlePabloMissing(context);
				}
			} else {
				await stopLanguageClientIfRunning();
				activeBinary = resolved;
				vscode.window.showWarningMessage(
					`Pablo executable saved for CLI, but LSP requires 1.3+: ${resolved}`
				);
			}
		})
	);

	context.subscriptions.push(vscode.commands.registerCommand('pablo.check', () => {
		void runPabloCommand('check', true);
	}));

	context.subscriptions.push(vscode.commands.registerCommand('pablo.init', () => {
		void runPabloCommand('init');
	}));

	context.subscriptions.push(vscode.commands.registerCommand('pablo.run', () => {
		void runPabloRun();
	}));

	context.subscriptions.push(
		vscode.commands.registerCommand(
			'pablo.runWithArgs',
			(uri: string, runTarget: string) => {
				void executeRun(vscode.Uri.parse(uri).fsPath, runTarget);
			}
		)
	);

	context.subscriptions.push(
		vscode.workspace.onDidChangeConfiguration(async (e) => {
			if (e.affectsConfiguration('pablo.trace.server') && client) {
				const level = vscode.workspace.getConfiguration('pablo').get<string>('trace.server', 'off');
				client.setTrace(parseTraceLevel(level));
				traceOutputChannel.appendLine(`LSP trace level changed: ${level}`);
			}
			if (e.affectsConfiguration('pablo.path')) {
				const started = await tryResolveAndStart(context);
				if (!started) {
					await stopLanguageClientIfRunning();
					await handlePabloMissing(context);
				}
			}
		})
	);

	await ensureLanguageClientStarted(context);

	const updateBinary = activeBinary ?? (await resolvePabloBinary(context, outputChannel));
	if (updateBinary) {
		void notifyCliUpdateIfAvailable({
			binaryPath: normalizeExecutablePath(updateBinary),
			outputChannel,
			onBeforeUpdate: stopLanguageClientIfRunning,
			onAfterUpdate: async () => {
				const started = await tryResolveAndStart(context);
				if (!started) {
					await handlePabloMissing(context);
				}
			},
		});
	}
}

async function runPabloRun() {
	const editor = vscode.window.activeTextEditor;
	if (!editor || !editor.document.fileName.match(/pablo.*\.ya?ml/)) {
		vscode.window.showErrorMessage('Open a pablo.yaml file to run deployment.');
		return;
	}

	const fileUri = editor.document.uri;
	const result = await inspectManifest(extensionContext, outputChannel, client, fileUri);
	if (!result) {
		return;
	}
	if (result.profiles.length === 0) {
		vscode.window.showErrorMessage('No profiles found in manifest.');
		return;
	}

	const profilePick = await vscode.window.showQuickPick(
		result.profiles.map((profile) => ({
			label: profile.name,
			description: profile.type,
			profile,
		})),
		{ title: 'Select profile' }
	);
	if (!profilePick) {
		return;
	}

	const environments = profilePick.profile.environments;
	if (environments.length === 0) {
		vscode.window.showErrorMessage(`Profile '${profilePick.profile.name}' has no environments.`);
		return;
	}

	const envPick = await vscode.window.showQuickPick(environments, { title: 'Select environment' });
	if (!envPick) {
		return;
	}

	await executeRun(fileUri.fsPath, `${profilePick.profile.name}/${envPick}`);
}

async function executeRun(filePath: string, runTarget: string): Promise<void> {
	const binary = await resolvePabloBinary(extensionContext, outputChannel);
	if (!binary) {
		await handlePabloMissing(extensionContext);
		return;
	}

	const resolved = normalizeExecutablePath(binary);
	const args = ['run', '--verbose', '-f', filePath, runTarget];
	const line = buildTerminalCommand(resolved, args);
	const terminal = vscode.window.terminals.find((t) => t.name === 'Pablo CLI')
		|| vscode.window.createTerminal('Pablo CLI');
	terminal.show();
	terminal.sendText(line);
}

async function runPabloCommand(command: string, fileSpecific: boolean = false) {
	const editor = vscode.window.activeTextEditor;
	const args: string[] = [command];

	if (fileSpecific) {
		if (editor) {
			args.push('-f', editor.document.fileName);
		} else {
			vscode.window.showErrorMessage('No active YAML editor found.');
			return;
		}
	} else if (editor && editor.document.fileName.match(/pablo.*\.ya?ml/)) {
		args.push('-f', editor.document.fileName);
	}

	const binary = await resolvePabloBinary(extensionContext, outputChannel);
	if (!binary) {
		await handlePabloMissing(extensionContext);
		return;
	}

	const resolved = normalizeExecutablePath(binary);
	const line = buildTerminalCommand(resolved, args);
	const terminal = vscode.window.terminals.find(t => t.name === 'Pablo CLI') || vscode.window.createTerminal('Pablo CLI');
	terminal.show();
	terminal.sendText(line);
}



export function deactivate(): Thenable<void> | undefined {

	if (!client) {

		return undefined;

	}

	isRestarting = true;

	if (client.state === State.Running) {

		return client.stop();

	}

	return undefined;

}

