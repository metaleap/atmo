import * as vscode from 'vscode'

import * as vfs from './vfs'



export function activate(ctx: vscode.ExtensionContext) {
	ctx.subscriptions.push(vscode.workspace.registerFileSystemProvider(
		vfs.uriScheme,
		new vfs.FS(),
		{ isCaseSensitive: !process.platform.startsWith('win') }))
}

export function deactivate() { }
