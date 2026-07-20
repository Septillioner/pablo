using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Text.RegularExpressions;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell.Interop;
using Microsoft.VisualStudio.Shell;
using Pablo.VisualStudio.Options;

namespace Pablo.VisualStudio.Services
{
    public sealed class PabloExecutableService
    {
        private readonly PabloUserSettings _settings;
        private readonly AsyncPackage _package;

        public PabloExecutableService(AsyncPackage package)
        {
            _package = package;
            _settings = PabloUserSettings.Load();
        }

        public void Log(string message)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            PabloOutputWindow.WriteLine(message);
        }

        public string? GetSelectedExecutable()
        {
            return _settings.SelectedExecutable;
        }

        public void SetSelectedExecutable(string path)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            _settings.SelectedExecutable = NormalizePath(path);
            _settings.Save();
        }

        public void ClearSelectedExecutable()
        {
            _settings.SelectedExecutable = null;
            _settings.Save();
        }

        public string NormalizePath(string binary)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            var normalized = Path.GetFullPath(binary);
            if (File.Exists(normalized))
            {
                return normalized;
            }

            var solutionDir = GetSolutionDirectory();
            if (!string.IsNullOrEmpty(solutionDir))
            {
                var fromSolution = Path.GetFullPath(Path.Combine(solutionDir, binary));
                if (File.Exists(fromSolution))
                {
                    return fromSolution;
                }
            }

            return normalized;
        }

        public async Task<string?> ResolveBinaryAsync()
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();

            var stored = _settings.SelectedExecutable;
            if (!string.IsNullOrWhiteSpace(stored) && File.Exists(stored))
            {
                return NormalizePath(stored);
            }

            if (!string.IsNullOrWhiteSpace(stored))
            {
                Log($"Stored Pablo executable not found: {stored}");
            }

            var configured = PabloOptionsPage.Instance?.ExecutablePath?.Trim();
            if (!string.IsNullOrWhiteSpace(configured))
            {
                var configuredPath = ResolveSettingPath(configured);
                if (File.Exists(configuredPath))
                {
                    return NormalizePath(configuredPath);
                }

                Log($"Configured pablo.path not found: {configuredPath}");
            }

            var onPath = await FindOnPathAsync();
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            if (onPath.Count == 1)
            {
                return NormalizePath(onPath[0]);
            }

            if (onPath.Count > 1)
            {
                Log($"Multiple Pablo binaries on PATH ({onPath.Count}). Use Pablo: Select Executable.");
            }

            return null;
        }

        public async Task<bool> AssertLspSupportedAsync(string binaryPath)
        {
            var result = await RunCommandAsync(binaryPath, new[] { "help", "lsp" });
            if (result.ExitCode != 0)
            {
                return false;
            }

            if (Regex.IsMatch(result.Output, @"unknown (command|help topic)", RegexOptions.IgnoreCase))
            {
                return false;
            }

            return Regex.IsMatch(result.Output, @"language server|stdio", RegexOptions.IgnoreCase);
        }

        public async Task<string?> ProbeVersionAsync(string binaryPath)
        {
            var result = await RunCommandAsync(binaryPath, new[] { "version" });
            if (result.ExitCode != 0)
            {
                return null;
            }

            var match = Regex.Match(result.Output, @"Pablo Version:\s*(\S+)", RegexOptions.IgnoreCase);
            if (!match.Success)
            {
                match = Regex.Match(result.Output, @"\bv?(\d+\.\d+\.\d+)\b", RegexOptions.IgnoreCase);
            }

            return match.Success ? match.Groups[1].Value : null;
        }

        public IReadOnlyList<string> DiscoverPickerCandidates()
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            var seen = new HashSet<string>(StringComparer.OrdinalIgnoreCase);
            var candidates = new List<string>();
            var solutionDir = GetSolutionDirectory();

            if (!string.IsNullOrWhiteSpace(_settings.SelectedExecutable) && File.Exists(_settings.SelectedExecutable))
            {
                AddCandidate(seen, candidates, _settings.SelectedExecutable);
            }

            var configured = PabloOptionsPage.Instance?.ExecutablePath?.Trim();
            if (!string.IsNullOrWhiteSpace(configured))
            {
                var configuredPath = ResolveSettingPath(configured);
                if (File.Exists(configuredPath))
                {
                    AddCandidate(seen, candidates, configuredPath);
                }
            }

            if (!string.IsNullOrEmpty(solutionDir))
            {
                foreach (var candidate in WorkspacePickerCandidates(solutionDir))
                {
                    if (File.Exists(candidate))
                    {
                        AddCandidate(seen, candidates, candidate);
                    }
                }
            }

            return candidates;
        }

        public Task<IReadOnlyList<string>> FindOnPathAsync()
        {
            return Task.Run(FindOnPathCore);
        }

        private static IReadOnlyList<string> FindOnPathCore()
        {
            try
            {
                var whereExe = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.System), "where.exe");
                var startInfo = new ProcessStartInfo
                {
                    FileName = whereExe,
                    Arguments = "pablo",
                    RedirectStandardOutput = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                };

                using var process = Process.Start(startInfo);
                if (process == null)
                {
                    return Array.Empty<string>();
                }

                var stdout = process.StandardOutput.ReadToEnd();
                if (!process.WaitForExit(PabloConstants.ProbeTimeoutMs))
                {
                    process.Kill();
                    return Array.Empty<string>();
                }

                return stdout
                    .Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries)
                    .Select(line => line.Trim())
                    .Where(IsExecutableCandidate)
                    .ToList();
            }
            catch
            {
                return Array.Empty<string>();
            }
        }

        private static void AddCandidate(HashSet<string> seen, List<string> list, string candidate)
        {
            var normalized = Path.GetFullPath(candidate);
            if (seen.Add(normalized))
            {
                list.Add(normalized);
            }
        }

        private static bool IsExecutableCandidate(string filePath)
        {
            try
            {
                return File.Exists(filePath) && filePath.EndsWith(".exe", StringComparison.OrdinalIgnoreCase);
            }
            catch
            {
                return false;
            }
        }

        private static IEnumerable<string> WorkspacePickerCandidates(string workspaceFolder)
        {
            var candidates = new List<string>();
            var buildDir = Path.Combine(workspaceFolder, "build");
            var defaultBuild = Path.Combine(buildDir, "pablo.exe");
            if (File.Exists(defaultBuild))
            {
                candidates.Add(defaultBuild);
            }

            if (Directory.Exists(buildDir))
            {
                foreach (var name in Directory.GetFiles(buildDir, "pablo*"))
                {
                    if (IsExecutableCandidate(name))
                    {
                        candidates.Add(name);
                    }
                }
            }

            var releasesDir = Path.Combine(workspaceFolder, "dist", "releases");
            if (Directory.Exists(releasesDir))
            {
                foreach (var name in Directory.GetFiles(releasesDir, "pablo*"))
                {
                    candidates.Add(name);
                }
            }

            return candidates;
        }

        private string ResolveSettingPath(string rawPath)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            var solutionDir = GetSolutionDirectory() ?? string.Empty;
            return rawPath
                .Replace("${workspaceFolder}", solutionDir)
                .Replace("${workspaceRoot}", solutionDir)
                .Trim();
        }

        private string? GetSolutionDirectory()
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            var solution = _package.GetServiceAsync(typeof(SVsSolution)).Result as IVsSolution;
            if (solution == null)
            {
                return null;
            }

            solution.GetSolutionInfo(out string? solutionDirectory, out _, out _);
            return string.IsNullOrWhiteSpace(solutionDirectory) ? null : solutionDirectory.TrimEnd('\\');
        }

        public async Task<CommandResult> RunCommandAsync(string binary, IReadOnlyList<string> args, int timeoutMs = PabloConstants.ProbeTimeoutMs)
        {
            return await Task.Run(() =>
            {
                if (!IsExecutableCandidate(binary) && !File.Exists(binary))
                {
                    return new CommandResult(-1, string.Empty);
                }

                var startInfo = new ProcessStartInfo
                {
                    FileName = binary,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                };

                PabloProcessHelper.SetArguments(startInfo, args);

                using var process = Process.Start(startInfo);
                if (process == null)
                {
                    return new CommandResult(-1, string.Empty);
                }

                var output = process.StandardOutput.ReadToEnd() + process.StandardError.ReadToEnd();
                if (!process.WaitForExit(timeoutMs))
                {
                    process.Kill();
                    return new CommandResult(-1, output);
                }

                return new CommandResult(process.ExitCode, output);
            });
        }

        public readonly struct CommandResult
        {
            public CommandResult(int exitCode, string output)
            {
                ExitCode = exitCode;
                Output = output;
            }

            public int ExitCode { get; }
            public string Output { get; }
        }
    }
}
