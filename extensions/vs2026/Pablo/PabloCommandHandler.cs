using System;
using System.Collections.Generic;
using System.ComponentModel.Design;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;
using Pablo.VisualStudio.Dialogs;
using Pablo.VisualStudio.Services;

namespace Pablo.VisualStudio
{
    internal static class PabloCommandHandler
    {
        private static PabloPackage? _package;
        private static OleMenuCommandService? _commandService;

        private static object[]? _pendingRunWithArgs;

        public static void SetPendingRunWithArgs(object[] args)
        {
            _pendingRunWithArgs = args;
        }

        public static async Task InitializeAsync(PabloPackage package)
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            _package = package;
            _commandService = await package.GetServiceAsync(typeof(IMenuCommandService)) as OleMenuCommandService;
            if (_commandService == null)
            {
                return;
            }

            Register(PabloCommandIds.Check, RunCheckAsync);
            Register(PabloCommandIds.Init, RunInitAsync);
            Register(PabloCommandIds.Run, RunDeploymentPickerAsync);
            Register(PabloCommandIds.SelectExecutable, SelectExecutableAsync);
            Register(PabloCommandIds.RunWithArgs, RunWithArgsFromMenuAsync);
        }

        private static void Register(PabloCommandIds commandId, Func<Task> handler)
        {
            var command = new OleMenuCommand(
                (_, __) => ThreadHelper.JoinableTaskFactory.RunAsync(async () => await handler()),
                new CommandID(PabloGuids.CommandSetGuid, (int)commandId));
            _commandService!.AddCommand(command);
        }

        private static async Task RunCheckAsync()
        {
            var filePath = GetActivePabloFilePath();
            if (filePath == null)
            {
                ShowError("Open a pablo.yaml file to check.");
                return;
            }

            await PabloTerminalRunner.RunCliAsync(_package!.ExecutableService, new[] { "check", "-f", filePath });
        }

        private static async Task RunInitAsync()
        {
            var filePath = GetActivePabloFilePath();
            var args = new List<string> { "init" };
            if (filePath != null)
            {
                args.Add("-f");
                args.Add(filePath);
            }

            await PabloTerminalRunner.RunCliAsync(_package!.ExecutableService, args.ToArray());
        }

        private static async Task RunDeploymentPickerAsync()
        {
            var filePath = GetActivePabloFilePath();
            if (filePath == null)
            {
                ShowError("Open a pablo.yaml file to run deployment.");
                return;
            }

            var inspect = await _package!.InspectService.InspectManifestAsync(filePath);
            if (inspect == null || inspect.Profiles.Length == 0)
            {
                ShowError("No profiles found in manifest.");
                return;
            }

            var profile = ProfileEnvPickerDialog.PickProfile(inspect.Profiles);
            if (profile == null)
            {
                return;
            }

            if (profile.Environments.Length == 0)
            {
                ShowError($"Profile '{profile.Name}' has no environments.");
                return;
            }

            var environment = ProfileEnvPickerDialog.PickEnvironment(profile.Environments);
            if (environment == null)
            {
                return;
            }

            await PabloTerminalRunner.RunDeploymentAsync(
                _package.ExecutableService,
                filePath,
                $"{profile.Name}/{environment}");
        }

        private static async Task RunWithArgsFromMenuAsync()
        {
            var args = _pendingRunWithArgs;
            _pendingRunWithArgs = null;
            if (args == null || args.Length < 2)
            {
                return;
            }

            var uri = Convert.ToString(args[0]) ?? string.Empty;
            var runTarget = Convert.ToString(args[1]) ?? string.Empty;
            if (string.IsNullOrWhiteSpace(uri) || string.IsNullOrWhiteSpace(runTarget))
            {
                return;
            }

            var filePath = new Uri(uri).LocalPath;
            await PabloTerminalRunner.RunDeploymentAsync(_package!.ExecutableService, filePath, runTarget);
        }

        public static async Task SelectExecutableAsync()
        {
            var executableService = _package!.ExecutableService;
            var candidates = executableService.DiscoverPickerCandidates().ToList();
            foreach (var pathCandidate in await executableService.FindOnPathAsync())
            {
                if (!candidates.Contains(pathCandidate, StringComparer.OrdinalIgnoreCase))
                {
                    candidates.Add(pathCandidate);
                }
            }

            var selected = ExecutablePickerDialog.PickExecutable(candidates, executableService.GetSelectedExecutable());
            if (string.IsNullOrWhiteSpace(selected))
            {
                return;
            }

            var resolved = executableService.NormalizePath(selected);
            executableService.SetSelectedExecutable(resolved);

            if (await executableService.AssertLspSupportedAsync(resolved))
            {
                await Lsp.PabloLanguageClientHost.RestartAsync();
                VsShellUtilities.ShowMessageBox(
                    ServiceProvider.GlobalProvider,
                    $"Pablo executable: {resolved}",
                    "Pablo",
                    OLEMSGICON.OLEMSGICON_INFO,
                    OLEMSGBUTTON.OLEMSGBUTTON_OK,
                    OLEMSGDEFBUTTON.OLEMSGDEFBUTTON_FIRST);
            }
            else
            {
                VsShellUtilities.ShowMessageBox(
                    ServiceProvider.GlobalProvider,
                    $"Pablo executable saved for CLI, but LSP requires 1.3+: {resolved}",
                    "Pablo",
                    OLEMSGICON.OLEMSGICON_WARNING,
                    OLEMSGBUTTON.OLEMSGBUTTON_OK,
                    OLEMSGDEFBUTTON.OLEMSGDEFBUTTON_FIRST);
            }
        }

        public static async Task PromptMissingExecutableAsync(PabloExecutableService executableService)
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            var result = VsShellUtilities.ShowMessageBox(
                ServiceProvider.GlobalProvider,
                "Pablo CLI not found or does not support LSP (pablo lsp). Pablo 1.3+ is required.",
                "Pablo",
                OLEMSGICON.OLEMSGICON_WARNING,
                OLEMSGBUTTON.OLEMSGBUTTON_OKCANCEL,
                OLEMSGDEFBUTTON.OLEMSGDEFBUTTON_FIRST);

            if (result == PabloConstants.MessageBoxCancelResult)
            {
                return;
            }

            await SelectExecutableAsync();
        }

        private static string? GetActivePabloFilePath()
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            var monitorSelection = ServiceProvider.GlobalProvider.GetService(typeof(SVsShellMonitorSelection)) as IVsMonitorSelection;
            if (monitorSelection == null)
            {
                return null;
            }

            monitorSelection.GetCurrentElementValue(PabloConstants.DocumentFrameSelectionId, out var frameObj);
            if (frameObj is not IVsWindowFrame frame)
            {
                return null;
            }

            frame.GetProperty((int)__VSFPROPID.VSFPROPID_pszMkDocument, out var documentPathObj);
            var documentPath = documentPathObj as string;
            if (string.IsNullOrWhiteSpace(documentPath))
            {
                return null;
            }

            var fileName = Path.GetFileName(documentPath);
            if (!ContentType.PabloFileToContentTypeProvider.IsPabloManifestFileName(fileName))
            {
                return null;
            }

            return documentPath;
        }

        private static void ShowError(string message)
        {
            VsShellUtilities.ShowMessageBox(
                ServiceProvider.GlobalProvider,
                message,
                "Pablo",
                OLEMSGICON.OLEMSGICON_CRITICAL,
                OLEMSGBUTTON.OLEMSGBUTTON_OK,
                OLEMSGDEFBUTTON.OLEMSGDEFBUTTON_FIRST);
        }
    }
}
