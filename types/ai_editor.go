// MIT License
//
// Copyright (c) 2024 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package types

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AIEditor represents an AI editor / viewer
type AIEditor struct {
	App           *AppContext     // the underlying application context
	ChatEditor    *tview.TextArea // the chat editor TextArea
	ChatHistory   *tview.List     // the chat history
	CreateButton  *tview.Button   // the "create" button
	FileViewer    *tview.TextView // the viewer for file content
	InfoLeft      *tview.TextView // the last info
	isCreating    bool
	isResetting   bool
	isSending     bool
	left          *tview.Flex
	OnCreateClick func() error                   // the callback that is executed then "create" button is "clicked"/"pressed"
	OnResetClick  func() error                   // the callback that is executed then "reset" button is "clicked"/"pressed"
	OnSendClick   func(chatMessage string) error // the callback that is executed then "send" button is "clicked"/"pressed"
	ProjectUrl    string                         // the URL of the new project, which is also the module name
	ResetButton   *tview.Button                  // the "reset" button
	Root          tview.Primitive                // the root element in the UI
	SendButton    *tview.Button                  // the "send" button
	Tree          *tview.TreeView                // the file tree
	TreeNodes     []*AIEditorFileTreeNode        // all current file tree nodes
	UI            *tview.Application             // the UI / App
}

// AIEditorFileItem is a simple type to update the file tree view
type AIEditorFileItem struct {
	Content []byte // the content
	Name    string // the name/relative path of the file
}

// AIEditorFileTreeNode is a "real" element in the file tree
type AIEditorFileTreeNode struct {
	Content  []byte                  // content
	Children []*AIEditorFileTreeNode // the children
	Name     string                  // the name/relative path of the file
	Node     *tview.TreeNode         // the node in the view
	Parent   *AIEditorFileTreeNode   // the parent
	Type     string                  // the type: `dir`, `file` or `root`
}

func (e *AIEditor) handle_create_button_click() {
	handleButtonClick := e.OnCreateClick
	if handleButtonClick == nil {
		return // no handler set
	}

	e.isCreating = true
	e.update_button_disable_states()

	go func() {
		err := handleButtonClick()

		e.isCreating = false
		e.update_button_disable_states()

		if err != nil {
			e.show_error(fmt.Sprintf("Could not create: %s", err.Error()))
		}

		e.UI.Draw()
	}()
}

func (e *AIEditor) handle_reset_button_click() {
	handleButtonClick := e.OnResetClick
	if handleButtonClick == nil {
		return // no handler set
	}

	e.isResetting = true
	e.update_button_disable_states()

	go func() {
		err := handleButtonClick()

		e.isResetting = false
		e.update_button_disable_states()

		if err != nil {
			e.show_error(fmt.Sprintf("Could not reset: %s", err.Error()))
		}

		e.UI.Draw()
	}()
}

func (e *AIEditor) handle_send_button_click() {
	textToSend := strings.TrimSpace(e.ChatEditor.GetText())
	if textToSend == "" {
		return // nothing to send
	}

	handleButtonClick := e.OnSendClick
	if handleButtonClick == nil {
		return // no handler set
	}

	e.isSending = true
	e.update_button_disable_states()

	go func() {
		err := handleButtonClick(textToSend)

		e.isSending = false
		e.update_button_disable_states()

		if err != nil {
			e.show_error(fmt.Sprintf("Could not send chat message: %s", err.Error()))
		}

		e.UI.Draw()
	}()
}

func (e *AIEditor) init_chat_editor() *tview.TextArea {
	textArea := tview.NewTextArea().
		SetPlaceholder(" Enter your new chat message here ")
	textArea.SetBorder(true)

	textArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// TAB
			e.UI.SetFocus(e.ChatHistory)
			return nil
		} else if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModShift {
			// SHIFT + ENTER
			textArea.SetText(fmt.Sprintf("%s\n", textArea.GetText()), true)
			return nil
		} else if event.Key() == tcell.KeyEnter {
			// ENTER => send
			e.handle_send_button_click()
			return nil
		}
		return event
	})

	textArea.SetChangedFunc(func() {
		e.update_send_button_disabled_state()
	})

	e.ChatEditor = textArea

	return textArea
}

func (e *AIEditor) init_chat_history() *tview.List {
	chatHistory := tview.NewList().
		ShowSecondaryText(false)

	chatHistory.
		SetBorder(true)

	chatHistory.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// TAB
			e.UI.SetFocus(e.SendButton)
			return nil
		}
		if event.Key() == tcell.KeyRight {
			// right
			e.UI.SetFocus(e.SendButton)
			return nil
		}
		if event.Key() == tcell.KeyLeft {
			// left
			e.UI.SetFocus(e.ChatEditor)
			return nil
		}
		return event
	})

	e.ChatHistory = chatHistory

	return chatHistory
}

func (e *AIEditor) init_create_button() *tview.Button {
	createButton := tview.NewButton("Create")

	createButton.SetBorder(true)

	createButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// TAB
			e.UI.SetFocus(e.ResetButton)
			return nil
		}
		if event.Key() == tcell.KeyRight {
			// right
			e.UI.SetFocus(e.ResetButton)
			return nil
		}
		if event.Key() == tcell.KeyUp {
			// up
			e.UI.SetFocus(e.Tree)
			return nil
		}
		return event
	})

	createButton.SetSelectedFunc(func() {
		e.handle_create_button_click()
	})

	e.CreateButton = createButton

	return createButton
}

func (e *AIEditor) init_file_viewer() *tview.TextView {
	fileViewer := tview.NewTextView().
		SetDynamicColors(false) // TODO: implement later

	fileViewer.SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)

	fileViewer.SetWordWrap(false)

	fileViewer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			e.UI.SetFocus(e.ChatEditor)
			return nil
		}
		return event
	})

	e.FileViewer = fileViewer

	return fileViewer
}

func (e *AIEditor) init_left_infobox() *tview.TextView {
	infoLeft := tview.NewTextView()

	infoLeft.SetBorder(true)

	e.InfoLeft = infoLeft

	return infoLeft
}

func (e *AIEditor) init_reset_button() *tview.Button {
	resetButton := tview.NewButton("Reset")

	resetButton.SetBorder(true).
		SetBackgroundColor(tcell.ColorRed).
		SetTitleColor(tcell.ColorYellow)

	resetButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// TAB
			e.UI.SetFocus(e.FileViewer)
			return nil
		}
		if event.Key() == tcell.KeyLeft {
			// left
			e.UI.SetFocus(e.CreateButton)
			return nil
		}
		if event.Key() == tcell.KeyRight {
			// right
			e.UI.SetFocus(e.SendButton)
			return nil
		}
		if event.Key() == tcell.KeyUp {
			// up
			e.UI.SetFocus(e.Tree)
			return nil
		}
		return event
	})

	resetButton.SetSelectedFunc(func() {
		e.handle_reset_button_click()
	})

	e.ResetButton = resetButton

	return resetButton
}

func (e *AIEditor) init_root() *tview.Flex {
	// "create" & "reset" buttons
	leftButtonGroup := tview.NewFlex().
		AddItem(e.CreateButton, 0, 1, false).
		AddItem(e.ResetButton, 0, 1, false)

	// chat editor with history
	rightChatEditor := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(e.ChatEditor, 0, 1, true).
		AddItem(e.ChatHistory, 24, 1, false)

	// complete left side
	left := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.Tree, 0, 1, false).
		AddItem(e.InfoLeft, 0, 0, false).
		AddItem(leftButtonGroup, 3, 1, false)

	// complete right side
	right := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.FileViewer, 0, 1, false).
		AddItem(rightChatEditor, 7, 1, true).
		AddItem(e.SendButton, 3, 1, false)

	// whole UI
	root := tview.NewFlex().
		AddItem(left, 0, 1, false).
		AddItem(right, 0, 2, true)

	e.left = left
	e.Root = root

	e.update_info_left()

	return root
}

func (e *AIEditor) init_send_button() *tview.Button {
	sendButton := tview.NewButton("").
		SetDisabled(true)

	sendButton.SetBorder(true)

	sendButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			e.UI.SetFocus(e.Tree)
			return nil
		}
		if event.Key() == tcell.KeyLeft {
			e.UI.SetFocus(e.ResetButton)
			return nil
		}
		if event.Key() == tcell.KeyUp {
			e.UI.SetFocus(e.ChatHistory)
			return nil
		}
		return event
	})

	sendButton.SetSelectedFunc(func() {
		e.handle_send_button_click()
	})

	e.SendButton = sendButton

	e.update_send_button_disabled_state()

	return sendButton
}

func (e *AIEditor) init_tree() *tview.TreeView {
	tree := tview.NewTreeView()

	tree.SetTitle(" Files ")

	tree.SetChangedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // this node does nothing
		}

		fileNode := reference.(*AIEditorFileTreeNode)
		if fileNode.Type != "file" {
			return // only files
		}

		fileName := filepath.Base(fileNode.Name)
		fileContent := fileNode.Content

		e.update_file_viewer(fileName, fileContent)
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// TAB
			e.UI.SetFocus(e.CreateButton)
			return nil
		}
		return event
	})

	tree.SetBorder(true).
		SetTitle(" Files ").
		SetTitleAlign(tview.AlignCenter)

	e.Tree = tree

	return tree
}

func (e *AIEditor) is_busy() bool {
	return e.isCreating ||
		e.isResetting ||
		e.isSending
}

// NewAIEditor creates a new instance of an `AIEditor`
// as reference by needing an `AppContext` and the project URL / module name
func NewAIEditor(app *AppContext, projectUrl string) *AIEditor {
	ui := tview.NewApplication()

	e := &AIEditor{}
	e.App = app
	e.ProjectUrl = projectUrl
	e.TreeNodes = make([]*AIEditorFileTreeNode, 0)
	e.UI = ui

	e.init_chat_editor()
	e.init_chat_history()
	e.init_create_button()
	e.init_file_viewer()
	e.init_left_infobox()
	e.init_reset_button()
	e.init_send_button()
	e.init_tree()

	e.init_root()

	return e
}

func (e *AIEditor) show_error(message string) {
	e.InfoLeft.
		SetTextColor(tcell.ColorRed).
		SetText(message)

	e.update_info_left()
}

func (e *AIEditor) rebuild_file_tree() {
	root := tview.NewTreeNode(e.ProjectUrl).
		SetColor(tcell.ColorRed)

	e.Tree.SetRoot(root).
		SetCurrentNode(root)

	// A helper function which adds the files and directories of the given path
	// to the given target node.
	add := func(parentNode *tview.TreeNode, node *AIEditorFileTreeNode) {
		if node.Children == nil {
			return // nothing to do
		}

		for _, child := range node.Children {
			name := filepath.Base(child.Name)

			treeNode := tview.NewTreeNode(name).
				SetReference(child).
				SetSelectable(true)

			if child.Type == "dir" {
				// special color for directories
				treeNode.SetColor(tcell.ColorGreen)
			}

			parentNode.AddChild(treeNode)

			node.Node = treeNode
		}
	}

	// build tree
	{
		rootNode := &AIEditorFileTreeNode{
			Children: e.TreeNodes,
			Content:  []byte{},
			Name:     "",
			Node:     root,
			Type:     "root",
		}

		add(root, rootNode)
	}

	// If a directory was selected, open it.
	e.Tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // this node does nothing
		}

		children := node.GetChildren()
		if len(children) == 0 {
			fileNode := reference.(*AIEditorFileTreeNode)

			add(node, fileNode)
		} else {
			node.SetExpanded(!node.IsExpanded())
		}
	})
}

// e.Run() runs the underlying UI as fullscreen application
func (e *AIEditor) Run() error {
	return e.UI.
		SetRoot(e.Root, true).
		EnableMouse(true).
		EnablePaste(false).
		Run()
}

func sort_ai_editor_file_nodes(nodes []*AIEditorFileTreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Type == nodes[j].Type {
			// then by name
			return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
		}
		return nodes[i].Type == "dir" // first display the directories
	})

	for _, node := range nodes {
		if node.Children != nil {
			sort_ai_editor_file_nodes(node.Children)
		}
	}
}

func (e *AIEditor) StopWith(f func() error) error {
	var err error = nil

	if f != nil {
		e.UI.Suspend(func() {
			errChan := make(chan error)

			go func(err chan error) {
				err <- f()
			}(errChan)

			err = <-errChan
		})
	}

	if err == nil {
		e.UI.Stop()
	}

	return err
}

func (e *AIEditor) update_button_disable_states() {
	e.update_create_button_disabled_state()
	e.update_chat_editor_disabled_state()
	e.update_reset_button_disabled_state()
	e.update_send_button_disabled_state()
}

func (e *AIEditor) update_chat_editor_disabled_state() {
	isEditorDisabled := e.is_busy()

	e.ChatEditor.SetDisabled(isEditorDisabled)

	e.update_ui()
}

func (e *AIEditor) update_create_button_disabled_state() {
	isCreateButtonDisabled := e.isCreating ||
		e.isResetting ||
		e.isSending

	e.CreateButton.SetDisabled(isCreateButtonDisabled)

	e.update_ui()
}

func (e *AIEditor) update_file_viewer(name string, content []byte) {
	e.FileViewer.SetTitle(fmt.Sprintf(" %v ", name))

	viewerText := string(content)

	// TODO: currently this code does not work as expected, so implement later
	/*
		lexerName := strings.TrimSpace(
			strings.ToLower(name),
		)
		for {
			if strings.HasPrefix(lexerName, ".") {
				lexerName = strings.TrimSpace(lexerName[1:])
			} else {
				break
			}
		}

		lexer := lexers.Get(lexerName)
		if lexer == nil {
			lexer = lexers.Fallback
		}

		styleName := utils.GetBestChromaStyleName()

		style := styles.Get(styleName)
		if style == nil {
			style = styles.Fallback
		}

		formatterName := utils.GetBestChromaFormatterName()
		formatter := formatters.Get(formatterName)

		iterator, err := lexer.Tokenise(nil, viewerText)
		if err == nil {
			var highlightedCode bytes.Buffer
			err := formatter.Format(&highlightedCode, style, iterator)
			if err == nil {
				viewerText = highlightedCode.String()
			}
		}
	*/

	e.FileViewer.
		SetText(viewerText)
}

func (e *AIEditor) update_info_left() {
	text := strings.TrimSpace(e.InfoLeft.GetText(true))
	if text == "" {
		e.left.ResizeItem(e.InfoLeft, 0, 0)
	} else {
		e.left.ResizeItem(e.InfoLeft, 7, 1)
	}
}

func (e *AIEditor) update_reset_button_disabled_state() {
	isResetButtonDisabled := e.is_busy()

	e.ResetButton.SetDisabled(isResetButtonDisabled)

	e.update_ui()
}

func (e *AIEditor) update_send_button_disabled_state() {
	isSendButtonDisabled := true
	newLabel := "Send"

	if e.isSending {
		newLabel = "Sending ..."
	} else {
		isSendButtonDisabled = e.is_busy() ||
			e.OnSendClick == nil ||
			strings.TrimSpace(
				e.ChatEditor.GetText(),
			) == ""
	}

	e.SendButton.SetLabel(newLabel)
	e.SendButton.SetDisabled(isSendButtonDisabled)

	e.update_ui()
}

func (e *AIEditor) update_ui() {
	// e.UI.Draw()
}

func (e *AIEditor) UpdateFileTree(fileItems []AIEditorFileItem) []*AIEditorFileTreeNode {
	rootNodes := make(map[string]*AIEditorFileTreeNode)
	allNodes := make(map[string]*AIEditorFileTreeNode)

	for _, fileItem := range fileItems {
		dirParts := strings.Split(fileItem.Name, "/")
		currentPath := ""
		var parent *AIEditorFileTreeNode

		for i, part := range dirParts {
			currentPath = path.Join(currentPath, part)

			if node, exists := allNodes[currentPath]; exists {
				// already exist => continue
				parent = node
				continue
			}

			nodeType := "file"
			if i < len(dirParts)-1 {
				nodeType = "dir" // directory
			}
			newNode := &AIEditorFileTreeNode{
				Name:     part,
				Parent:   parent,
				Type:     nodeType,
				Children: []*AIEditorFileTreeNode{},
			}

			if nodeType == "file" {
				// assign content
				newNode.Content = fileItem.Content
			}

			if parent != nil {
				parent.Children = append(parent.Children, newNode) // child
			} else {
				rootNodes[currentPath] = newNode // root
			}

			// track node globally
			allNodes[currentPath] = newNode
			parent = newNode
		}
	}

	// map => slice
	roots := make([]*AIEditorFileTreeNode, 0)
	for _, node := range rootNodes {
		roots = append(roots, node)
	}

	// sort recursivly
	sort_ai_editor_file_nodes(roots)

	// set new data and update view
	e.TreeNodes = roots
	e.rebuild_file_tree()

	return roots
}
