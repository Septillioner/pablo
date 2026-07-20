using System;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;
using Newtonsoft.Json;
using Pablo.VisualStudio.Lsp;

namespace Pablo.VisualStudio.Services
{
    internal static class PabloUpdateService
    {
        private static readonly string[] UpdateCheckArgs = { "update", "check", "--json" };
        private static readonly string[] UpdateInstallArgs = { "update" };

        private sealed class UpdateCheckJson
        {
            [JsonProperty("current_version")]
            public string? CurrentVersion { get; set; }

            [JsonProperty("latest_version")]
            public string? LatestVersion { get; set; }

            [JsonProperty("release_tag")]
            public string? ReleaseTag { get; set; }

            [JsonProperty("update_available")]
            public bool? UpdateAvailable { get; set; }
        }

        /// <summary>Activation path: silent when up to date; prompt only when an update is available.</summary>
        public static Task CheckAndNotifyAsync(PabloExecutableService executableService)
        {
            return RunUpdateFlowAsync(executableService, interactive: false);
        }

        /// <summary>Manual "Pablo: Update" command: always report status to the user.</summary>
        public static Task RunUpdateCommandAsync(PabloExecutableService executableService)
        {
            return RunUpdateFlowAsync(executableService, interactive: true);
        }

        private static async Task RunUpdateFlowAsync(
            PabloExecutableService executableService,
            bool interactive)
        {
            var binary = await executableService.ResolveBinaryAsync();
            if (string.IsNullOrWhiteSpace(binary))
            {
                if (interactive)
                {
                    await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
                    ShowMessage(
                        "Pablo CLI not found. Use Pablo: Select Executable or set the Pablo path in Options.",
                        OLEMSGICON.OLEMSGICON_WARNING);
                }

                return;
            }

            PabloExecutableService.CommandResult probe;
            try
            {
                probe = await executableService.RunCommandAsync(
                    binary,
                    UpdateCheckArgs,
                    PabloConstants.UpdateCheckTimeoutMs);
            }
            catch (Exception ex)
            {
                PabloOutputWindow.WriteLine($"CLI update check failed: {ex.Message}");
                if (interactive)
                {
                    await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
                    ShowMessage(
                        $"Could not check for Pablo CLI updates: {ex.Message}",
                        OLEMSGICON.OLEMSGICON_WARNING);
                }

                return;
            }

            var result = ParseUpdateCheckJson(probe.Output);
            if (result == null)
            {
                PabloOutputWindow.WriteLine("CLI update check skipped (offline, unsupported CLI, or parse error).");
                if (interactive)
                {
                    await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
                    ShowMessage(
                        "Could not check for Pablo CLI updates. You may be offline, or this CLI does not support update check.",
                        OLEMSGICON.OLEMSGICON_WARNING);
                }

                return;
            }

            if (!result.UpdateAvailable)
            {
                var upToDate = $"Pablo CLI is up to date ({result.CurrentVersion}).";
                PabloOutputWindow.WriteLine(upToDate);
                if (interactive)
                {
                    await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
                    ShowMessage(upToDate, OLEMSGICON.OLEMSGICON_INFO);
                }

                return;
            }

            var summary = $"Pablo CLI update available: {result.CurrentVersion} -> {result.LatestVersion}";
            PabloOutputWindow.WriteLine(summary);

            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            var choice = VsShellUtilities.ShowMessageBox(
                ServiceProvider.GlobalProvider,
                summary + "\n\nUpdate now?",
                "Pablo",
                OLEMSGICON.OLEMSGICON_INFO,
                OLEMSGBUTTON.OLEMSGBUTTON_YESNO,
                OLEMSGDEFBUTTON.OLEMSGDEFBUTTON_FIRST);

            if (choice != PabloConstants.MessageBoxYesResult)
            {
                return;
            }

            try
            {
                await PabloLanguageClientHost.StopAsync();

                var install = await executableService.RunCommandAsync(
                    binary,
                    UpdateInstallArgs,
                    PabloConstants.UpdateInstallTimeoutMs);

                if (install.ExitCode != 0)
                {
                    PabloOutputWindow.WriteLine($"CLI update failed:\n{install.Output}");
                    ShowMessage(
                        "Pablo CLI update failed. Check Output > Pablo Language Server for details.",
                        OLEMSGICON.OLEMSGICON_CRITICAL);
                    return;
                }

                PabloOutputWindow.WriteLine($"CLI update succeeded:\n{install.Output}");
                ShowMessage(
                    $"Pablo CLI updated to {result.LatestVersion}.",
                    OLEMSGICON.OLEMSGICON_INFO);
            }
            finally
            {
                await PabloLanguageClientHost.RestartAsync();
            }
        }

        private static void ShowMessage(string message, OLEMSGICON icon)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            VsShellUtilities.ShowMessageBox(
                ServiceProvider.GlobalProvider,
                message,
                "Pablo",
                icon,
                OLEMSGBUTTON.OLEMSGBUTTON_OK,
                OLEMSGDEFBUTTON.OLEMSGDEFBUTTON_FIRST);
        }

        private static UpdateCheckResult? ParseUpdateCheckJson(string raw)
        {
            if (string.IsNullOrWhiteSpace(raw))
            {
                return null;
            }

            var trimmed = raw.Trim();
            var jsonStart = trimmed.IndexOf('{');
            var jsonEnd = trimmed.LastIndexOf('}');
            if (jsonStart < 0 || jsonEnd <= jsonStart)
            {
                return null;
            }

            try
            {
                var parsed = JsonConvert.DeserializeObject<UpdateCheckJson>(
                    trimmed.Substring(jsonStart, jsonEnd - jsonStart + 1));
                if (parsed?.UpdateAvailable == null)
                {
                    return null;
                }

                return new UpdateCheckResult(
                    parsed.CurrentVersion ?? string.Empty,
                    parsed.LatestVersion ?? string.Empty,
                    parsed.ReleaseTag ?? string.Empty,
                    parsed.UpdateAvailable.Value);
            }
            catch
            {
                return null;
            }
        }

        private sealed class UpdateCheckResult
        {
            public UpdateCheckResult(string currentVersion, string latestVersion, string releaseTag, bool updateAvailable)
            {
                CurrentVersion = currentVersion;
                LatestVersion = latestVersion;
                ReleaseTag = releaseTag;
                UpdateAvailable = updateAvailable;
            }

            public string CurrentVersion { get; }
            public string LatestVersion { get; }
            public string ReleaseTag { get; }
            public bool UpdateAvailable { get; }
        }
    }
}
