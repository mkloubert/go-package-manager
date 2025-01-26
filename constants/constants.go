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

package constants

// AI APIs
const AIApiOllama = "ollama"
const AIApiOpenAI = "openai"

// operating system
const DefaultDirMode = 0750
const DefaultFileMode = 0750
const WindowsExecutableExt = ".exe"

// source
const DefaultAliasSource = "https://raw.githubusercontent.com/mkloubert/go-package-manager/refs/heads/main/aliases.yaml"
const DefaultProjectSource = "https://raw.githubusercontent.com/mkloubert/go-package-manager/refs/heads/main/projects.yaml"

// scripts
const BumpScriptName = "bump"
const PostBumpScriptName = "postbump"
const PostInstallScriptName = "postinstall"
const PostPackScriptName = "postpack"
const PostPublishScriptName = "postpublish"
const PostTestScriptName = "test"
const PostTidyScriptName = "posttidy"
const PreBumpScriptName = "prebump"
const PreInstallScriptName = "preinstall"
const PrePackScriptName = "prepack"
const PrePublishScriptName = "prepublish"
const PreStartScriptName = "prestart"
const PreTestScriptName = "test"
const PreTidyScriptName = "pretidy"
const PublishScriptName = "publish"
const StartScriptName = "start"
const TestScriptName = "test"
const TidyScriptName = "tidy"
