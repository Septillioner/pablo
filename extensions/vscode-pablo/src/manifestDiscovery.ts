import * as path from 'path';
import * as vscode from 'vscode';

const MANIFEST_GLOB_EXACT_YAML = '**/pablo.yaml';
const MANIFEST_GLOB_EXACT_YML = '**/pablo.yml';
const MANIFEST_GLOB_PREFIX_YAML = '**/pablo*.yaml';
const MANIFEST_GLOB_PREFIX_YML = '**/pablo*.yml';
const MANIFEST_EXCLUDE_GLOB = '{**/node_modules/**,**/.git/**}';
const EXACT_MANIFEST_BASENAMES = new Set(['pablo.yaml', 'pablo.yml']);

export function isExactManifestName(filePath: string): boolean {
	return EXACT_MANIFEST_BASENAMES.has(path.basename(filePath).toLowerCase());
}

function compareManifestUris(a: vscode.Uri, b: vscode.Uri): number {
	const aExact = isExactManifestName(a.fsPath) ? 0 : 1;
	const bExact = isExactManifestName(b.fsPath) ? 0 : 1;
	if (aExact !== bExact) {
		return aExact - bExact;
	}
	return a.fsPath.localeCompare(b.fsPath);
}

export async function discoverManifestUris(): Promise<vscode.Uri[]> {
	const [exactYaml, exactYml, prefixYaml, prefixYml] = await Promise.all([
		vscode.workspace.findFiles(MANIFEST_GLOB_EXACT_YAML, MANIFEST_EXCLUDE_GLOB),
		vscode.workspace.findFiles(MANIFEST_GLOB_EXACT_YML, MANIFEST_EXCLUDE_GLOB),
		vscode.workspace.findFiles(MANIFEST_GLOB_PREFIX_YAML, MANIFEST_EXCLUDE_GLOB),
		vscode.workspace.findFiles(MANIFEST_GLOB_PREFIX_YML, MANIFEST_EXCLUDE_GLOB),
	]);

	const byPath = new Map<string, vscode.Uri>();
	for (const uri of [...exactYaml, ...exactYml, ...prefixYaml, ...prefixYml]) {
		byPath.set(uri.fsPath, uri);
	}

	return [...byPath.values()].sort(compareManifestUris);
}

export function formatManifestLabel(uri: vscode.Uri): string {
	const folder = vscode.workspace.getWorkspaceFolder(uri);
	const relative = folder
		? vscode.workspace.asRelativePath(uri, false)
		: uri.fsPath;
	const folders = vscode.workspace.workspaceFolders;
	if (folders && folders.length > 1 && folder) {
		return path.posix.join(folder.name, relative.replace(/\\/g, '/'));
	}
	return relative.replace(/\\/g, '/');
}
