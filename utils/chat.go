package utils

// ChatPromptSuggestion stores data for an element that is returned
// by GetChatPromptSugesstions()
type ChatPromptSuggestion struct {
	Description string // the long description
	Text        string // the short description of the command
}

// GetChatPromptSugesstions() - returns a list of commands and their descriptions for the AI chat
// in a very generic and platform independent form
func GetChatPromptSugesstions() []ChatPromptSuggestion {
	return []ChatPromptSuggestion{
		{Text: "/cls", Description: "clear screen"},
		{Text: "/exit", Description: "exit application"},
		{Text: "/format <name>", Description: "formatter for console output"},
		{Text: "/info", Description: "print information about current chat settings and status"},
		{Text: "/model <name>", Description: "switch to another model"},
		{Text: "/nosystem", Description: "delete system prompt"},
		{Text: "/reset", Description: "reset conversation"},
		{Text: "/style <name>", Description: "console style"},
		{Text: "/system <text>", Description: "reset conversation and update system prompt"},
		{Text: "/temp <value>", Description: "custom temperature value"},
	}
}
