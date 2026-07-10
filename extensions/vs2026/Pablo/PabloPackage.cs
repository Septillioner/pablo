using System;
using System.Runtime.InteropServices;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;
using Pablo.VisualStudio.Options;
using Pablo.VisualStudio.Services;
using Pablo.VisualStudio.ToolWindows;

namespace Pablo.VisualStudio
{
    [PackageRegistration(UseManagedResourcesOnly = true, AllowsBackgroundLoading = true)]
    [ProvideMenuResource("Menus.ctmenu", 1)]
    [ProvideToolWindow(typeof(PabloDeployToolWindow))]
    [ProvideOptionPage(typeof(PabloOptionsPage), "Pablo", "General", 0, 0, true)]
    [ProvideAutoLoad(UIContextGuids80.SolutionExists, PackageAutoLoadFlags.BackgroundLoad)]
    [Guid(PabloGuids.PackageGuidString)]
    public sealed class PabloPackage : AsyncPackage
    {
        public static PabloPackage Instance { get; private set; } = null!;

        public PabloExecutableService ExecutableService { get; private set; } = null!;
        public PabloInspectService InspectService { get; private set; } = null!;

        protected override async Task InitializeAsync(CancellationToken cancellationToken, IProgress<ServiceProgressData> progress)
        {
            await JoinableTaskFactory.SwitchToMainThreadAsync(cancellationToken);
            Instance = this;
            ExecutableService = new PabloExecutableService(this);
            InspectService = new PabloInspectService(ExecutableService);

            PabloOutputWindow.WriteLine("Pablo extension is now active.");
            await PabloCommandHandler.InitializeAsync(this);
            PabloToolbarController.Instance.RefreshManifests();
            await Lsp.PabloLanguageClientHost.RestartAsync();
        }

        internal void OnOptionsChanged()
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            _ = JoinableTaskFactory.RunAsync(async () =>
            {
                await Lsp.PabloLanguageClientHost.RestartAsync();
            });
        }
    }
}
