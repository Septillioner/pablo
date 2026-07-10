const { spawnSync } = require('child_process');
const path = require('path');

const pkg = require('../package.json');
const vsixFile = path.join(__dirname, '..', `pablo-${pkg.version}.vsix`);
const ovsxCli = require.resolve('ovsx/bin/ovsx');

const result = spawnSync(process.execPath, [ovsxCli, 'publish', vsixFile], {
	stdio: 'inherit',
	env: process.env,
});

if (result.error) {
	throw result.error;
}

process.exit(result.status ?? 1);
