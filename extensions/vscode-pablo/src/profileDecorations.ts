import * as vscode from 'vscode';

/** Hue-spread palette for profiles (document order). */
const PROFILE_COLORS = [
	'#5C6BC0',
	'#26A69A',
	'#EF5350',
	'#AB47BC',
	'#42A5F5',
	'#66BB6A',
];

/** Hue-spread palette for environments (document order). */
const ENV_COLORS = [
	'#E53935',
	'#1E88E5',
	'#43A047',
	'#FB8C00',
	'#8E24AA',
	'#00ACC1',
	'#FDD835',
	'#6D4C41',
	'#EC407A',
	'#3949AB',
	'#00897B',
	'#C0CA33',
];

const DEBOUNCE_MS = 150;
const BORDER_WIDTH_PX = 3;

interface MarkedLine {
	line: number;
	colorKey: string;
	indent: number;
	kind: 'profile' | 'env';
}

const decorationCache = new Map<string, { color: string; type: vscode.TextEditorDecorationType }>();
let debounceTimer: ReturnType<typeof setTimeout> | undefined;

function isPabloManifest(document: vscode.TextDocument): boolean {
	return document.languageId === 'pablo' || /pablo.*\.ya?ml$/i.test(document.fileName);
}

function getIndent(line: string): number {
	const match = line.match(/^(\s*)/);
	return match ? match[1].length : 0;
}

function getMappingKey(line: string): string | undefined {
	const match = line.match(/^\s*([a-zA-Z0-9_-]+):\s*/);
	return match?.[1];
}

function uniqueColorKeysInOrder(marked: MarkedLine[], kind: 'profile' | 'env'): string[] {
	const keys: string[] = [];
	const seen = new Set<string>();
	for (const entry of marked) {
		if (entry.kind !== kind || seen.has(entry.colorKey)) {
			continue;
		}
		seen.add(entry.colorKey);
		keys.push(entry.colorKey);
	}
	return keys;
}

function assignPaletteColors(colorKeys: string[], palette: string[]): Map<string, string> {
	const assigned = new Map<string, string>();
	colorKeys.forEach((key, index) => {
		assigned.set(key, palette[index % palette.length]);
	});
	return assigned;
}

function getDecorationForColor(colorKey: string, color: string): vscode.TextEditorDecorationType {
	const cached = decorationCache.get(colorKey);
	if (cached && cached.color === color) {
		return cached.type;
	}
	if (cached) {
		cached.type.dispose();
	}

	const type = vscode.window.createTextEditorDecorationType({
		borderWidth: `0 0 0 ${BORDER_WIDTH_PX}px`,
		borderStyle: 'solid',
		borderColor: color,
		borderRadius: '0',
	});
	decorationCache.set(colorKey, { color, type });
	return type;
}

/**
 * Paint continuous profile stripe through the whole profile subtree,
 * plus a second stripe at env indent for each environment block.
 */
export function parseProfileMarkedLines(text: string): MarkedLine[] {
	const lines = text.split(/\r?\n/);
	const marked: MarkedLine[] = [];
	let profilesIndent = -1;
	let currentProfile = '';
	let profileIndent = -1;
	let environmentsIndent = -1;
	let currentEnv = '';
	let envIndent = -1;

	for (let lineNum = 0; lineNum < lines.length; lineNum++) {
		const line = lines[lineNum];
		const trimmed = line.trim();
		const indent = getIndent(line);
		const key = getMappingKey(line);

		if (key === 'profiles') {
			profilesIndent = indent;
			currentProfile = '';
			profileIndent = -1;
			environmentsIndent = -1;
			currentEnv = '';
			envIndent = -1;
			continue;
		}

		if (profilesIndent < 0) {
			continue;
		}

		if (trimmed && indent <= profilesIndent) {
			currentProfile = '';
			profileIndent = -1;
			environmentsIndent = -1;
			currentEnv = '';
			envIndent = -1;
			continue;
		}

		if (key && indent === profilesIndent + 2) {
			currentProfile = key;
			profileIndent = indent;
			environmentsIndent = -1;
			currentEnv = '';
			envIndent = -1;
			marked.push({
				line: lineNum,
				colorKey: `profile:${currentProfile}`,
				indent: profileIndent,
				kind: 'profile',
			});
			continue;
		}

		if (!currentProfile || profileIndent < 0) {
			continue;
		}

		if (trimmed && indent <= profileIndent) {
			currentProfile = '';
			profileIndent = -1;
			environmentsIndent = -1;
			currentEnv = '';
			envIndent = -1;
			continue;
		}

		if (key === 'environments' && indent > profileIndent) {
			environmentsIndent = indent;
			currentEnv = '';
			envIndent = -1;
			marked.push({
				line: lineNum,
				colorKey: `profile:${currentProfile}`,
				indent: profileIndent,
				kind: 'profile',
			});
			continue;
		}

		if (
			key
			&& environmentsIndent >= 0
			&& indent === environmentsIndent + 2
		) {
			currentEnv = key;
			envIndent = indent;
			marked.push({
				line: lineNum,
				colorKey: `profile:${currentProfile}`,
				indent: profileIndent,
				kind: 'profile',
			});
			marked.push({
				line: lineNum,
				colorKey: `env:${currentProfile}/${currentEnv}`,
				indent: envIndent,
				kind: 'env',
			});
			continue;
		}

		if (currentEnv && envIndent >= 0 && trimmed && indent <= envIndent) {
			currentEnv = '';
			envIndent = -1;
		}

		if (environmentsIndent >= 0 && trimmed && indent <= environmentsIndent && key) {
			environmentsIndent = -1;
			currentEnv = '';
			envIndent = -1;
		}

		// Continuous profile stripe for every line under the profile.
		marked.push({
			line: lineNum,
			colorKey: `profile:${currentProfile}`,
			indent: profileIndent,
			kind: 'profile',
		});

		// Env stripe only while inside an environment block.
		if (currentEnv && envIndent >= 0 && (trimmed === '' || indent > envIndent)) {
			marked.push({
				line: lineNum,
				colorKey: `env:${currentProfile}/${currentEnv}`,
				indent: envIndent,
				kind: 'env',
			});
		}
	}

	return marked;
}

function lineDecorationRange(
	document: vscode.TextDocument,
	line: number,
	indent: number
): vscode.Range {
	const text = document.lineAt(line).text;
	const lineLength = text.length;

	// Empty lines: no reliable character cell — skip by using a no-op range at 0
	// that still gets a left border when the line has been padded by the editor.
	if (lineLength === 0) {
		return new vscode.Range(line, 0, line, 0);
	}

	// Paint the whitespace cell just before this indent so the stripe sits in
	// the gutter space and does not cover the first character of the key.
	const end = Math.min(Math.max(indent, 1), lineLength);
	const start = Math.max(0, end - 1);
	return new vscode.Range(line, start, line, end);
}

function applyDecorations(editor: vscode.TextEditor): void {
	if (!isPabloManifest(editor.document)) {
		return;
	}

	const marked = parseProfileMarkedLines(editor.document.getText());
	const profileKeys = uniqueColorKeysInOrder(marked, 'profile');
	const envKeys = uniqueColorKeysInOrder(marked, 'env');
	const palette = new Map<string, string>([
		...assignPaletteColors(profileKeys, PROFILE_COLORS),
		...assignPaletteColors(envKeys, ENV_COLORS),
	]);

	const rangesByColorKey = new Map<string, vscode.Range[]>();
	for (const entry of marked) {
		const ranges = rangesByColorKey.get(entry.colorKey) ?? [];
		ranges.push(lineDecorationRange(editor.document, entry.line, entry.indent));
		rangesByColorKey.set(entry.colorKey, ranges);
	}

	for (const colorKey of decorationCache.keys()) {
		if (!rangesByColorKey.has(colorKey)) {
			const cached = decorationCache.get(colorKey);
			if (cached) {
				editor.setDecorations(cached.type, []);
				cached.type.dispose();
				decorationCache.delete(colorKey);
			}
		}
	}

	// Profile first, then env — env stripe sits to the right of profile stripe.
	for (const kind of ['profile', 'env'] as const) {
		for (const [colorKey, ranges] of rangesByColorKey) {
			if (!colorKey.startsWith(`${kind}:`)) {
				continue;
			}
			const color = palette.get(colorKey);
			if (!color) {
				continue;
			}
			editor.setDecorations(getDecorationForColor(colorKey, color), ranges);
		}
	}
}

function scheduleUpdate(editor?: vscode.TextEditor): void {
	if (debounceTimer) {
		clearTimeout(debounceTimer);
	}
	debounceTimer = setTimeout(() => {
		const activeEditor = editor ?? vscode.window.activeTextEditor;
		if (activeEditor) {
			applyDecorations(activeEditor);
		}
	}, DEBOUNCE_MS);
}

export function registerProfileDecorations(context: vscode.ExtensionContext): void {
	context.subscriptions.push(
		vscode.window.onDidChangeActiveTextEditor((editor) => {
			if (editor) {
				scheduleUpdate(editor);
			}
		}),
		vscode.workspace.onDidOpenTextDocument((document) => {
			if (isPabloManifest(document) && vscode.window.activeTextEditor?.document === document) {
				scheduleUpdate();
			}
		}),
		vscode.workspace.onDidChangeTextDocument((event) => {
			if (isPabloManifest(event.document) && vscode.window.activeTextEditor?.document === event.document) {
				scheduleUpdate();
			}
		}),
		{
			dispose: () => {
				if (debounceTimer) {
					clearTimeout(debounceTimer);
				}
				for (const cached of decorationCache.values()) {
					cached.type.dispose();
				}
				decorationCache.clear();
			},
		}
	);

	if (vscode.window.activeTextEditor) {
		scheduleUpdate(vscode.window.activeTextEditor);
	}
}
