import { execFile } from 'child_process';
import { promisify } from 'util';
import * as vscode from 'vscode';
import { LanguageClient, State } from 'vscode-languageclient/node';

import { normalizeExecutablePath, resolvePabloBinary } from './executable';

const execFileAsync = promisify(execFile);

export interface InspectProfile {
	name: string;
	type: string;
	environments: string[];
}

export interface InspectResult {
	name: string;
	version: string;
	profiles: InspectProfile[];
}

async function inspectViaCLI(binary: string, filePath: string): Promise<InspectResult> {
	const resolved = normalizeExecutablePath(binary);
	const { stdout } = await execFileAsync(resolved, ['inspect', '-f', filePath, '--json']);
	return JSON.parse(stdout) as InspectResult;
}

export async function inspectManifest(
	context: vscode.ExtensionContext,
	outputChannel: vscode.OutputChannel,
	client: LanguageClient | undefined,
	fileUri: vscode.Uri
): Promise<InspectResult | undefined> {
	if (client?.state === State.Running) {
		try {
			const result = await client.sendRequest<InspectResult>('pablo/listProfiles', {
				uri: fileUri.toString(),
			});
			if (result?.profiles) {
				return result;
			}
		} catch (err: unknown) {
			const message = err instanceof Error ? err.message : String(err);
			outputChannel.appendLine(`LSP listProfiles failed, falling back to CLI: ${message}`);
		}
	}

	const binary = await resolvePabloBinary(context, outputChannel);
	if (!binary) {
		return undefined;
	}

	try {
		return await inspectViaCLI(binary, fileUri.fsPath);
	} catch (err: unknown) {
		const message = err instanceof Error ? err.message : String(err);
		outputChannel.appendLine(`CLI inspect failed: ${message}`);
		vscode.window.showErrorMessage(`Failed to inspect manifest: ${message}`);
		return undefined;
	}
}
