using System;
using System.Collections.Generic;
using System.ComponentModel.Design;
using System.IO;
using System.Linq;
using System.Runtime.InteropServices;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;
using Pablo.VisualStudio.Dialogs;
using Pablo.VisualStudio.Services;
using Pablo.VisualStudio.ToolWindows;

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
            Register(PabloCommandIds.Run, ShowDeployToolWindowAsync);
            Register(PabloCommandIds.ShowDeploy, ShowDeployToolWindowAsync);
            Register(PabloCommandIds.SelectExecutable, SelectExecutableAsync);
            Register(PabloCommandIds.RunWithArgs, RunWithArgsFromMenuAsync);
            Register(PabloCommandIds.ToolbarRun, RunFromToolbarAsync);

            RegisterCombo(PabloCommandIds.YamlCombo, OnYamlComboInvoke);
            RegisterComboGetList(PabloCommandIds.YamlComboGetList, OnYamlComboGetList);
            RegisterCombo(PabloCommandIds.ProfileCombo, OnProfileComboInvoke);
            RegisterComboGetList(PabloCommandIds.ProfileComboGetList, OnProfileComboGetList);
            RegisterCombo(PabloCommandIds.EnvCombo, OnEnvComboInvoke);
            RegisterComboGetList(PabloCommandIds.EnvComboGetList, OnEnvComboGetList);
        }

        private static void RegisterCombo(PabloCommandIds commandId, EventHandler handler)
        {
            var command = new OleMenuCommand(handler, new CommandID(PabloGuids.CommandSetGuid, (int)commandId));
            _commandService!.AddCommand(command);
        }

        private static void RegisterComboGetList(PabloCommandIds commandId, EventHandler handler)
        {
            var command = new OleMenuCommand(handler, new CommandID(PabloGuids.CommandSetGuid, (int)commandId));
            _commandService!.AddCommand(command);
        }

        private static void OnYamlComboInvoke(object sender, EventArgs e)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            if (e is OleMenuCmdEventArgs args)
            {
                HandleComboInvoke(
                    args,
                    () => PabloToolbarController.Instance.CurrentManifestLabel,
                    PabloToolbarController.Instance.SelectManifestByLabel);
            }
        }

        private static void OnYamlComboGetList(object sender, EventArgs e)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            if (e is OleMenuCmdEventArgs args)
            {
                PabloToolbarController.Instance.RefreshManifests();
                SetComboList(args, PabloToolbarController.Instance.GetManifestLabels());
            }
        }

        private static void OnProfileComboInvoke(object sender, EventArgs e)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            if (e is OleMenuCmdEventArgs args)
            {
                HandleComboInvoke(
                    args,
                    () => PabloToolbarController.Instance.CurrentProfileLabel,
                    PabloToolbarController.Instance.SelectProfileByLabel);
            }
        }

        private static void OnProfileComboGetList(object sender, EventArgs e)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            if (e is OleMenuCmdEventArgs args)
            {
                PabloToolbarController.Instance.EnsureInspectForSelectedManifest();
                SetComboList(args, PabloToolbarController.Instance.GetProfileLabels());
            }
        }

        private static void OnEnvComboInvoke(object sender, EventArgs e)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            if (e is OleMenuCmdEventArgs args)
            {
                HandleComboInvoke(
                    args,
                    () => PabloToolbarController.Instance.CurrentEnvironmentLabel,
                    PabloToolbarController.Instance.SelectEnvironmentByLabel);
            }
        }

        private static void OnEnvComboGetList(object sender, EventArgs e)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            if (e is OleMenuCmdEventArgs args)
            {
                PabloToolbarController.Instance.EnsureInspectForSelectedManifest();
                SetComboList(args, PabloToolbarController.Instance.GetEnvironmentLabels());
            }
        }

        private static void HandleComboInvoke(
            OleMenuCmdEventArgs args,
            Func<string> getCurrentLabel,
            Action<string> selectByLabel)
        {
            if (args.OutValue != IntPtr.Zero)
            {
                Marshal.GetNativeVariantForObject(getCurrentLabel() ?? string.Empty, args.OutValue);
                return;
            }

            if (args.InValue is string label && !string.IsNullOrWhiteSpace(label))
            {
                selectByLabel(label);
            }
        }

        private static void SetComboList(OleMenuCmdEventArgs args, IReadOnlyList<string> labels)
        {
            if (args.OutValue == IntPtr.Zero)
            {
                return;
            }

            var array = labels.Count == 0 ? Array.Empty<string>() : labels.ToArray();
            Marshal.GetNativeVariantForObject(array, args.OutValue);
        }

        private static async Task RunFromToolbarAsync()
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            var ran = await PabloToolbarController.Instance.RunSelectedDeploymentAsync();
            if (!ran)
            {
                ShowError("Select a manifest, profile, and environment on the Pablo toolbar.");
            }
        }

        private static void Register(PabloCommandIds commandId, Func<Task> handler)
        {
            var command = new OleMenuCommand(
                (_, __) => ThreadHelper.JoinableTaskFactory.RunAsync(async () => await handler()),
                new CommandID(PabloGuids.CommandSetGuid, (int)commandId));
            _commandService!.AddCommand(command);
        }

        public static async Task ShowDeployToolWindowAsync()
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            if (_package == null)
            {
                return;
            }

            if (await _package.ShowToolWindowAsync(
                    typeof(PabloDeployToolWindow),
                    0,
                    create: true,
                    cancellationToken: CancellationToken.None) is PabloDeployToolWindow toolWindow)
            {
                await toolWindow.RefreshAsync();
            }
        }

        private static async Task RunCheckAsync()
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            var filePath = PabloManifestPathResolver.ResolveActiveManifestPath();
            if (filePath == null)
            {
                ShowError("Open a pablo.yaml file to check.");
                return;
            }

            await PabloTerminalRunner.RunCliAsync(_package!.ExecutableService, new[] { "check", "-f", filePath });
        }

        private static async Task RunInitAsync()
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            var filePath = PabloManifestPathResolver.ResolveActiveManifestPath();
            var args = new List<string> { "init" };
            if (filePath != null)
            {
                args.Add("-f");
                args.Add(filePath);
            }

            await PabloTerminalRunner.RunCliAsync(_package!.ExecutableService, args.ToArray());
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
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
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

            var version = await executableService.ProbeVersionAsync(resolved);
            var versionLabel = string.IsNullOrWhiteSpace(version) ? "unknown" : version;
            var lspSupported = await executableService.AssertLspSupportedAsync(resolved);

            if (lspSupported)
            {
                await Lsp.PabloLanguageClientHost.RestartAsync();
                VsShellUtilities.ShowMessageBox(
                    ServiceProvider.GlobalProvider,
                    $"Pablo executable: {resolved}\nVersion: {versionLabel}\nLSP: supported",
                    "Pablo",
                    OLEMSGICON.OLEMSGICON_INFO,
                    OLEMSGBUTTON.OLEMSGBUTTON_OK,
                    OLEMSGDEFBUTTON.OLEMSGDEFBUTTON_FIRST);
            }
            else
            {
                VsShellUtilities.ShowMessageBox(
                    ServiceProvider.GlobalProvider,
                    $"Pablo executable saved for CLI: {resolved}\nVersion: {versionLabel}\nLSP: not supported (this binary has no `pablo lsp`; use Pablo 1.3+).",
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
