using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;
using Pablo.VisualStudio.Options;

namespace Pablo.VisualStudio.Services
{
    internal static class PabloBinaryResolver
    {
        public static async Task<string?> ResolveBinaryAsync(PabloExecutableService? preferredService = null)
        {
            if (preferredService != null)
            {
                var fromService = await preferredService.ResolveBinaryAsync();
                if (!string.IsNullOrWhiteSpace(fromService))
                {
                    return fromService;
                }
            }

            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();

            var settings = PabloUserSettings.Load();
            if (!string.IsNullOrWhiteSpace(settings.SelectedExecutable) && File.Exists(settings.SelectedExecutable))
            {
                return Path.GetFullPath(settings.SelectedExecutable);
            }

            var configured = PabloOptionsPage.Instance?.ExecutablePath?.Trim();
            if (!string.IsNullOrWhiteSpace(configured))
            {
                var configuredPath = ExpandSettingPath(configured);
                if (File.Exists(configuredPath))
                {
                    return Path.GetFullPath(configuredPath);
                }
            }

            var onPath = await FindOnPathAsync();
            if (onPath.Count == 1)
            {
                return onPath[0];
            }

            return null;
        }

        public static async Task<bool> AssertLspSupportedAsync(string binaryPath)
        {
            return await Task.Run(() =>
            {
                if (!File.Exists(binaryPath))
                {
                    return false;
                }

                var startInfo = new ProcessStartInfo
                {
                    FileName = binaryPath,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                };
                PabloProcessHelper.SetArguments(startInfo, new[] { "help", "lsp" });

                using var process = Process.Start(startInfo);
                if (process == null)
                {
                    return false;
                }

                var output = process.StandardOutput.ReadToEnd() + process.StandardError.ReadToEnd();
                if (!process.WaitForExit(PabloConstants.ProbeTimeoutMs))
                {
                    process.Kill();
                    return false;
                }

                if (process.ExitCode != 0)
                {
                    return false;
                }

                if (Regex.IsMatch(output, @"unknown (command|help topic)", RegexOptions.IgnoreCase))
                {
                    return false;
                }

                return Regex.IsMatch(output, @"language server|stdio", RegexOptions.IgnoreCase);
            });
        }

        private static string ExpandSettingPath(string rawPath)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            var solutionDir = TryGetSolutionDirectory() ?? string.Empty;
            return rawPath
                .Replace("${workspaceFolder}", solutionDir)
                .Replace("${workspaceRoot}", solutionDir)
                .Trim();
        }

        private static string? TryGetSolutionDirectory()
        {
            try
            {
                ThreadHelper.ThrowIfNotOnUIThread();
                var package = PabloPackage.Instance;
                if (package == null)
                {
                    return null;
                }

                var solution = package.GetServiceAsync(typeof(SVsSolution)).GetAwaiter().GetResult() as IVsSolution;
                if (solution == null)
                {
                    return null;
                }

                solution.GetSolutionInfo(out string? solutionDirectory, out _, out _);
                return string.IsNullOrWhiteSpace(solutionDirectory) ? null : solutionDirectory.TrimEnd('\\');
            }
            catch
            {
                return null;
            }
        }

        private static Task<IReadOnlyList<string>> FindOnPathAsync()
        {
            return Task.Run(() =>
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
                        return (IReadOnlyList<string>)Array.Empty<string>();
                    }

                    var stdout = process.StandardOutput.ReadToEnd();
                    if (!process.WaitForExit(PabloConstants.ProbeTimeoutMs))
                    {
                        process.Kill();
                        return (IReadOnlyList<string>)Array.Empty<string>();
                    }

                    return stdout
                        .Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries)
                        .Select(line => line.Trim())
                        .Where(path => path.EndsWith(".exe", StringComparison.OrdinalIgnoreCase) && File.Exists(path))
                        .Select(Path.GetFullPath)
                        .ToList();
                }
                catch
                {
                    return (IReadOnlyList<string>)Array.Empty<string>();
                }
            });
        }
    }
}
