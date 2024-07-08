package session

import (
	"os"
	"path/filepath"
	"strings"

	"atmo/util"
	"atmo/util/sl"
)

var (
	allSrcFiles map[string]*SrcFile

	OnNoticesChanged func(map[string][]*SrcFileNotice)
)

func init() {
	allSrcFiles = map[string]*SrcFile{}
}

type SrcFile struct {
	FilePath string
	Content  struct {
		Src                string
		TopLevelToksChunks ToksChunks
		TopLevelAstNodes   Nodes
	}
	Notices struct {
		LastReadErr *SrcFileNotice
		LexErrs     []*SrcFileNotice
		ParseErrs   []*SrcFileNotice
	}
}

func OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string) {
	for _, file_path := range removed {
		delete(allSrcFiles, file_path)
	}
	for _, file_path := range current {
		EnsureSrcFile(file_path, nil, canSkipFileRead)
	}
	refreshAndPublishNotices(append(removed, current...)...)
}

func OnSrcFileEdit(srcFilePath string, curFullContent string) {
	EnsureSrcFile(srcFilePath, &curFullContent, true)
	refreshAndPublishNotices(srcFilePath)
}

func EnsureSrcFile(srcFilePath string, curFullContent *string, canSkipFileRead bool) *SrcFile {
	util.Assert(IsSrcFilePath(srcFilePath), srcFilePath)
	src_file := allSrcFiles[srcFilePath]
	if src_file == nil {
		src_file = &SrcFile{FilePath: srcFilePath}
		allSrcFiles[srcFilePath] = src_file
	}
	old_content, had_last_read_err := src_file.Content.Src, (src_file.Notices.LastReadErr != nil)
	if curFullContent != nil {
		src_file.Content.Src, src_file.Notices.LastReadErr = *curFullContent, nil
	} else if (!canSkipFileRead) || had_last_read_err || (old_content == "") {
		src_file_bytes, err := os.ReadFile(srcFilePath)
		src_file.Content.Src, src_file.Notices.LastReadErr = string(src_file_bytes), errToNotice(err, NoticeCodeFileReadError)
	}
	if (src_file.Content.Src != old_content) || had_last_read_err || (src_file.Notices.LastReadErr != nil) {
		src_file.Content.TopLevelAstNodes, src_file.Content.TopLevelToksChunks, src_file.Notices.LexErrs, src_file.Notices.ParseErrs =
			nil, nil, nil, nil
		if src_file.Notices.LastReadErr == nil {
			src_file.Content.TopLevelToksChunks, src_file.Notices.LexErrs = tokenize(src_file.Content.Src, srcFilePath)
			top_level_nodes := src_file.Content.TopLevelAstNodes
			var gone_nodes Nodes
			src_file.Notices.ParseErrs = nil
			// remove nodes whose src is no longer present
			for i := 0; i < len(top_level_nodes); i++ {
				if !sl.Any(src_file.Content.TopLevelToksChunks, func(topLevelChunk Toks) bool {
					return topLevelChunk.src(src_file.Content.Src) == top_level_nodes[i].ToksSrc
				}) {
					gone_nodes = append(gone_nodes, top_level_nodes[i])
					top_level_nodes = append(top_level_nodes[:i], top_level_nodes[i+1:]...)
					i--
				}
			}
			// parse only top-level chunks whose nodes do not exist
			var new_nodes Nodes
			for _, top_level_chunk := range src_file.Content.TopLevelToksChunks {
				node := sl.FirstWhere(top_level_nodes, func(it *Node) bool { return it.ToksSrc == top_level_chunk.src(src_file.Content.Src) })
				if node == nil {
					node, err := src_file.parseNode(top_level_chunk)
					if err != nil {
						src_file.Notices.ParseErrs = append(src_file.Notices.ParseErrs, err)
					} else {
						new_nodes = append(new_nodes, node)
					}
				}
			}
			// if despite changed tokens, the parsed node has existed before, keep it (to keep annotations etc)
			for i := 0; i < len(new_nodes); i++ {
				new_node := new_nodes[i]
				if old_node := sl.FirstWhere(gone_nodes, func(it *Node) bool { return it.equals(new_node) }); old_node != nil {
					old_node.Toks, old_node.ToksSrc = new_node.Toks, new_node.ToksSrc
					gone_nodes = sl.Without(gone_nodes, true, old_node)
					new_nodes = append(new_nodes[:i], append(Nodes{old_node}, new_nodes[i+1:]...)...)
					i--
				}
			}
			top_level_nodes = append(top_level_nodes, new_nodes...)
			// TODO: sort all nodes to be in source-file order of appearance

			src_file.Content.TopLevelAstNodes = top_level_nodes
		}
	}
	return src_file
}

func IsSrcFilePath(filePath string) bool {
	return filepath.IsAbs(filePath) && filepath.Ext(filePath) == ".at" &&
		(!strings.Contains(filePath, string(filepath.Separator)+".")) && (!util.FsIsDir(filePath))
}
