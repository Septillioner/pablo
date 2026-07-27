import * as vscode from 'vscode';
import { LanguageClient } from 'vscode-languageclient/node';

import { InspectProfile, inspectManifest } from './inspect';
import {
	discoverManifestUris,
	formatManifestLabel,
	isExactManifestName,
} from './manifestDiscovery';

export const DEPLOY_VIEW_ID = 'pablo.deployView';

export interface DeploySelection {
	filePath: string;
	profile: string;
	environment: string;
}

export interface DeployViewDeps {
	context: vscode.ExtensionContext;
	outputChannel: vscode.OutputChannel;
	getClient: () => LanguageClient | undefined;
	onRun: (filePath: string, runTarget: string) => Promise<void>;
}

interface ManifestOption {
	path: string;
	label: string;
}

interface WebviewState {
	manifests: ManifestOption[];
	selectedFile: string;
	profiles: Array<{ name: string; type: string }>;
	selectedProfile: string;
	environments: string[];
	selectedEnvironment: string;
	status: string;
	canRun: boolean;
}

type WebviewMessage =
	| { type: 'ready' }
	| { type: 'refresh' }
	| { type: 'selectFile'; path: string }
	| { type: 'selectProfile'; name: string }
	| { type: 'selectEnvironment'; name: string }
	| { type: 'run' };

export class DeployViewProvider implements vscode.WebviewViewProvider {
	private view?: vscode.WebviewView;
	private manifests: ManifestOption[] = [];
	private selectedFile = '';
	private profiles: InspectProfile[] = [];
	private selectedProfile = '';
	private environments: string[] = [];
	private selectedEnvironment = '';
	private status = 'Loading manifests...';
	private refreshInFlight?: Promise<void>;

	constructor(private readonly deps: DeployViewDeps) {}

	getSelection(): DeploySelection | undefined {
		if (!this.selectedFile || !this.selectedProfile || !this.selectedEnvironment) {
			return undefined;
		}
		return {
			filePath: this.selectedFile,
			profile: this.selectedProfile,
			environment: this.selectedEnvironment,
		};
	}

	getSelectedFilePath(): string | undefined {
		return this.selectedFile || undefined;
	}

	resolveWebviewView(
		webviewView: vscode.WebviewView,
		_context: vscode.WebviewViewResolveContext,
		_token: vscode.CancellationToken
	): void {
		this.view = webviewView;
		webviewView.webview.options = {
			enableScripts: true,
		};
		webviewView.webview.html = this.getHtml(webviewView.webview);

		webviewView.webview.onDidReceiveMessage((message: WebviewMessage) => {
			void this.handleMessage(message);
		});

		webviewView.onDidDispose(() => {
			if (this.view === webviewView) {
				this.view = undefined;
			}
		});

		void this.refresh();
	}

	async refresh(): Promise<void> {
		if (this.refreshInFlight) {
			return this.refreshInFlight;
		}

		this.refreshInFlight = this.refreshInternal().finally(() => {
			this.refreshInFlight = undefined;
		});
		return this.refreshInFlight;
	}

	private async refreshInternal(): Promise<void> {
		this.status = 'Loading manifests...';
		this.postState();

		const uris = await discoverManifestUris();
		this.manifests = uris.map((uri) => ({
			path: uri.fsPath,
			label: formatManifestLabel(uri),
		}));

		if (this.manifests.length === 0) {
			this.selectedFile = '';
			this.profiles = [];
			this.selectedProfile = '';
			this.environments = [];
			this.selectedEnvironment = '';
			this.status = 'No pablo.yaml or pablo*.yaml found in the workspace.';
			this.postState();
			return;
		}

		const stillValid = this.manifests.some((item) => item.path === this.selectedFile);
		if (!stillValid) {
			const preferred =
				this.manifests.find((item) => isExactManifestName(item.path))
				?? this.manifests[0];
			this.selectedFile = preferred.path;
		}

		await this.loadProfilesForSelectedFile();
	}

	private async loadProfilesForSelectedFile(): Promise<void> {
		if (!this.selectedFile) {
			this.profiles = [];
			this.selectedProfile = '';
			this.environments = [];
			this.selectedEnvironment = '';
			this.status = 'Select a manifest file.';
			this.postState();
			return;
		}

		this.status = 'Inspecting manifest...';
		this.postState();

		const fileUri = vscode.Uri.file(this.selectedFile);
		const result = await inspectManifest(
			this.deps.context,
			this.deps.outputChannel,
			this.deps.getClient(),
			fileUri
		);

		if (!result) {
			this.profiles = [];
			this.selectedProfile = '';
			this.environments = [];
			this.selectedEnvironment = '';
			this.status = 'Failed to inspect manifest. Check Output > Pablo Language Server.';
			this.postState();
			return;
		}

		this.profiles = result.profiles;
		if (this.profiles.length === 0) {
			this.selectedProfile = '';
			this.environments = [];
			this.selectedEnvironment = '';
			this.status = 'No profiles found in manifest.';
			this.postState();
			return;
		}

		if (!this.profiles.some((profile) => profile.name === this.selectedProfile)) {
			this.selectedProfile = this.profiles[0].name;
		}

		this.applyProfileEnvironments();
		this.status = '';
		this.postState();
	}

	private applyProfileEnvironments(): void {
		const profile = this.profiles.find((item) => item.name === this.selectedProfile);
		this.environments = profile?.environments ?? [];
		if (!this.environments.includes(this.selectedEnvironment)) {
			this.selectedEnvironment = this.environments[0] ?? '';
		}
	}

	private async handleMessage(message: WebviewMessage): Promise<void> {
		switch (message.type) {
			case 'ready':
				this.postState();
				return;
			case 'refresh':
				await this.refresh();
				return;
			case 'selectFile':
				if (message.path === this.selectedFile) {
					return;
				}
				this.selectedFile = message.path;
				this.selectedProfile = '';
				this.selectedEnvironment = '';
				await this.loadProfilesForSelectedFile();
				return;
			case 'selectProfile':
				if (message.name === this.selectedProfile) {
					return;
				}
				this.selectedProfile = message.name;
				this.applyProfileEnvironments();
				this.postState();
				return;
			case 'selectEnvironment':
				this.selectedEnvironment = message.name;
				this.postState();
				return;
			case 'run':
				await this.runSelected();
				return;
		}
	}

	private async runSelected(): Promise<void> {
		const selection = this.getSelection();
		if (!selection) {
			vscode.window.showErrorMessage('Select a manifest, profile, and environment before running.');
			return;
		}
		await this.deps.onRun(selection.filePath, `${selection.profile}/${selection.environment}`);
	}

	private buildState(): WebviewState {
		return {
			manifests: this.manifests,
			selectedFile: this.selectedFile,
			profiles: this.profiles.map((profile) => ({
				name: profile.name,
				type: profile.type,
			})),
			selectedProfile: this.selectedProfile,
			environments: this.environments,
			selectedEnvironment: this.selectedEnvironment,
			status: this.status,
			canRun: Boolean(this.selectedFile && this.selectedProfile && this.selectedEnvironment),
		};
	}

	private postState(): void {
		void this.view?.webview.postMessage({
			type: 'state',
			state: this.buildState(),
		});
	}

	private getHtml(webview: vscode.Webview): string {
		const csp = [
			`default-src 'none'`,
			`style-src ${webview.cspSource} 'unsafe-inline'`,
			`script-src 'nonce-pabloDeploy'`,
		].join('; ');

		return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<meta http-equiv="Content-Security-Policy" content="${csp}" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<style>
		:root {
			color-scheme: light dark;
		}
		body {
			font-family: var(--vscode-font-family);
			font-size: var(--vscode-font-size);
			color: var(--vscode-foreground);
			background: transparent;
			padding: 12px;
			margin: 0;
		}
		h1 {
			font-size: 11px;
			font-weight: 600;
			letter-spacing: 0.04em;
			text-transform: uppercase;
			margin: 0 0 12px;
			opacity: 0.75;
		}
		label {
			display: block;
			margin: 12px 0 4px;
			font-size: 11px;
			opacity: 0.8;
		}
		select, button {
			width: 100%;
			box-sizing: border-box;
			font: inherit;
		}
		select {
			background: var(--vscode-dropdown-background);
			color: var(--vscode-dropdown-foreground);
			border: 1px solid var(--vscode-dropdown-border);
			padding: 6px 8px;
		}
		select:disabled {
			opacity: 0.55;
		}
		.actions {
			display: flex;
			gap: 8px;
			margin-top: 16px;
		}
		button {
			background: var(--vscode-button-background);
			color: var(--vscode-button-foreground);
			border: none;
			padding: 8px 10px;
			cursor: pointer;
		}
		button.secondary {
			background: var(--vscode-button-secondaryBackground);
			color: var(--vscode-button-secondaryForeground);
		}
		button:hover:not(:disabled) {
			background: var(--vscode-button-hoverBackground);
		}
		button.secondary:hover:not(:disabled) {
			background: var(--vscode-button-secondaryHoverBackground);
		}
		button:disabled {
			opacity: 0.5;
			cursor: default;
		}
		.status {
			margin-top: 12px;
			font-size: 12px;
			opacity: 0.75;
			min-height: 1.2em;
			white-space: pre-wrap;
		}
	</style>
</head>
<body>
	<h1>Deploy</h1>
	<label for="manifest">Manifest</label>
	<select id="manifest"></select>
	<label for="profile">Profile</label>
	<select id="profile"></select>
	<label for="environment">Environment</label>
	<select id="environment"></select>
	<div class="actions">
		<button id="run" type="button">Run Deployment</button>
		<button id="refresh" class="secondary" type="button">Refresh</button>
	</div>
	<div id="status" class="status"></div>
	<script nonce="pabloDeploy">
		const vscode = acquireVsCodeApi();
		const manifestEl = document.getElementById('manifest');
		const profileEl = document.getElementById('profile');
		const environmentEl = document.getElementById('environment');
		const runEl = document.getElementById('run');
		const refreshEl = document.getElementById('refresh');
		const statusEl = document.getElementById('status');

		function fillSelect(el, items, selected, getValue, getLabel) {
			el.innerHTML = '';
			if (!items.length) {
				const option = document.createElement('option');
				option.value = '';
				option.textContent = 'None';
				el.appendChild(option);
				el.disabled = true;
				return;
			}
			el.disabled = false;
			for (const item of items) {
				const option = document.createElement('option');
				option.value = getValue(item);
				option.textContent = getLabel(item);
				if (option.value === selected) {
					option.selected = true;
				}
				el.appendChild(option);
			}
		}

		function render(state) {
			fillSelect(
				manifestEl,
				state.manifests,
				state.selectedFile,
				(item) => item.path,
				(item) => item.label
			);
			fillSelect(
				profileEl,
				state.profiles,
				state.selectedProfile,
				(item) => item.name,
				(item) => item.type ? (item.name + ' (' + item.type + ')') : item.name
			);
			fillSelect(
				environmentEl,
				state.environments.map((name) => ({ name })),
				state.selectedEnvironment,
				(item) => item.name,
				(item) => item.name
			);
			runEl.disabled = !state.canRun;
			statusEl.textContent = state.status || '';
		}

		manifestEl.addEventListener('change', () => {
			vscode.postMessage({ type: 'selectFile', path: manifestEl.value });
		});
		profileEl.addEventListener('change', () => {
			vscode.postMessage({ type: 'selectProfile', name: profileEl.value });
		});
		environmentEl.addEventListener('change', () => {
			vscode.postMessage({ type: 'selectEnvironment', name: environmentEl.value });
		});
		runEl.addEventListener('click', () => {
			vscode.postMessage({ type: 'run' });
		});
		refreshEl.addEventListener('click', () => {
			vscode.postMessage({ type: 'refresh' });
		});

		window.addEventListener('message', (event) => {
			const message = event.data;
			if (message && message.type === 'state') {
				render(message.state);
			}
		});

		vscode.postMessage({ type: 'ready' });
	</script>
</body>
</html>`;
	}
}

export function registerDeployView(deps: DeployViewDeps): DeployViewProvider {
	const provider = new DeployViewProvider(deps);
	deps.context.subscriptions.push(
		vscode.window.registerWebviewViewProvider(DEPLOY_VIEW_ID, provider)
	);

	const watcher = vscode.workspace.createFileSystemWatcher('**/pablo*.{yaml,yml}');
	const scheduleRefresh = () => {
		void provider.refresh();
	};
	watcher.onDidCreate(scheduleRefresh);
	watcher.onDidDelete(scheduleRefresh);
	watcher.onDidChange(scheduleRefresh);
	deps.context.subscriptions.push(watcher);

	deps.context.subscriptions.push(
		vscode.workspace.onDidChangeWorkspaceFolders(() => {
			void provider.refresh();
		})
	);

	return provider;
}
