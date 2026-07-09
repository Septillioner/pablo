using System;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;

namespace Pablo.VisualStudio
{
    internal static class PabloOutputWindow
    {
        private static IVsOutputWindowPane? _pane;

        public static void WriteLine(string message)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            EnsurePane();
            _pane?.OutputStringThreadSafe($"{DateTime.Now:HH:mm:ss} {message}{Environment.NewLine}");
        }

        private static void EnsurePane()
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            if (_pane != null)
            {
                return;
            }

            var outputWindow = ServiceProvider.GlobalProvider.GetService(typeof(SVsOutputWindow)) as IVsOutputWindow;
            if (outputWindow == null)
            {
                return;
            }

            var paneGuid = new Guid("f3a8b2c1-4d5e-6f70-8a9b-0c1d2e3f4a5b");
            outputWindow.CreatePane(ref paneGuid, "Pablo Language Server", 1, 1);
            outputWindow.GetPane(ref paneGuid, out _pane);
        }
    }
}
