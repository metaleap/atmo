import * as node_path from 'path'
import * as vscode from 'vscode'

export const uriScheme = "atmo"

export const fileExts = [".go.at", ".ts.at"]


function isAtmoFile(uri: vscode.Uri) {
  return fileExts.some(ext => uri.path.endsWith(ext))
}

function uriWith(uri: vscode.Uri, keepPath = false, scheme = 'file') {
  return uri.with({ scheme: scheme, path: keepPath || !isAtmoFile(uri) ? uri.path : uri.path.substring(0, uri.path.length - ".at".length) })
}


export class File implements vscode.FileStat {
  type: vscode.FileType
  ctime: number
  mtime: number
  size: number

  name: string
  data?: Uint8Array

  constructor(name: string) {
    this.type = vscode.FileType.File
    this.ctime = Date.now()
    this.mtime = Date.now()
    this.size = 0
    this.name = name
  }
}
export class Directory implements vscode.FileStat {
  type: vscode.FileType
  ctime: number
  mtime: number
  size: number

  name: string
  entries: Map<string, File | Directory>

  constructor(name: string) {
    this.type = vscode.FileType.Directory
    this.ctime = Date.now()
    this.mtime = Date.now()
    this.size = 0
    this.name = name
    this.entries = new Map()
  }
}

export type Entry = File | Directory

export class FS implements vscode.FileSystemProvider {
  subs = new Map<string, FileWatchOptions>()

  emitter = new vscode.EventEmitter<vscode.FileChangeEvent[]>();

  onDidChangeFile: vscode.Event<vscode.FileChangeEvent[]> = this.emitter.event;

  async copy?(uriSrc: vscode.Uri, uriDst: vscode.Uri, options: { readonly overwrite: boolean }): Promise<void> {
    uriSrc = uriWith(uriSrc, (await vscode.workspace.fs.stat(uriWith(uriSrc, true))).type === vscode.FileType.Directory)
    uriDst = uriWith(uriDst, (await vscode.workspace.fs.stat(uriWith(uriDst, true))).type === vscode.FileType.Directory)
    return vscode.workspace.fs.copy(uriSrc, uriDst, options)
  }

  async createDirectory(uri: vscode.Uri): Promise<void> {
    uri = uriWith(uri, true)
    await vscode.workspace.fs.createDirectory(uri)
    this.emitter.fire([{ type: vscode.FileChangeType.Created, uri: uri }])
  }

  async delete(uri: vscode.Uri, options: { readonly recursive: boolean }): Promise<void> {
    const isDir = (await vscode.workspace.fs.stat(uriWith(uri, true))).type === vscode.FileType.Directory
    uri = uriWith(uri, isDir)
    await vscode.workspace.fs.delete(uri, options)
  }

  readDirectory(uri: vscode.Uri): [string, vscode.FileType][] | Thenable<[string, vscode.FileType][]> {
    uri = uriWith(uri, true)
    return vscode.workspace.fs.readDirectory(uri)
  }

  readFile(uri: vscode.Uri): Uint8Array | Thenable<Uint8Array> {
    uri = uriWith(uri)
    return vscode.workspace.fs.readFile(uri)
  }

  async rename(uriOld: vscode.Uri, uriNew: vscode.Uri, options: { readonly overwrite: boolean }): Promise<void> {
    uriOld = uriWith(uriOld, (await vscode.workspace.fs.stat(uriWith(uriOld, true))).type === vscode.FileType.Directory)
    uriNew = uriWith(uriNew, (await vscode.workspace.fs.stat(uriWith(uriNew, true))).type === vscode.FileType.Directory)
    return vscode.workspace.fs.rename(uriOld, uriNew, options)
  }

  stat(uri: vscode.Uri): vscode.FileStat | Thenable<vscode.FileStat> {
    uri = uriWith(uri, true)
    return vscode.workspace.fs.stat(uri)
  }

  writeFile(uri: vscode.Uri, content: Uint8Array, options: { readonly create: boolean; readonly overwrite: boolean }): void | Thenable<void> {
    uri = uriWith(uri)
    return vscode.workspace.fs.writeFile(uri, content)
  }

  watch(uri: vscode.Uri, options: FileWatchOptions): vscode.Disposable {
    uri = uri.with({ scheme: 'file' })
    this.subs.set(uri.toString(), options)
    return new vscode.Disposable(() => { this.subs.delete(uri.toString()) })
  }

}

type FileWatchOptions = { readonly recursive: boolean; readonly excludes: readonly string[] }
