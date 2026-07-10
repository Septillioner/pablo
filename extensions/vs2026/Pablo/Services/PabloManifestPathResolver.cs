using System;
using System.IO;
using EnvDTE;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;
using Pablo.VisualStudio.ContentType;

namespace Pablo.VisualStudio.Services
{
    internal static class PabloManifestPathResolver
    {
        public static string? ResolveActiveManifestPath()
        {
            ThreadHelper.ThrowIfNotOnUIThread();

            var fromDte = TryGetFromDte();
            if (fromDte != null)
            {
                return fromDte;
            }

            var fromFrame = TryGetFromDocumentFrame();
            if (fromFrame != null)
            {
                return fromFrame;
            }

            var fromTracker = PabloManifestTracker.LastKnownPath;
            if (fromTracker != null)
            {
                return fromTracker;
            }

            return TryGetFromOpenDocuments();
        }

        public static string? NormalizeIfPabloManifest(string? path)
        {
            if (string.IsNullOrWhiteSpace(path))
            {
                return null;
            }

            var localPath = NormalizeToLocalPath(path);
            if (!IsPabloManifestFileName(Path.GetFileName(localPath)))
            {
                return null;
            }

            return localPath;
        }

        public static bool IsPabloManifestFileName(string fileName)
        {
            return PabloFileToContentTypeProvider.IsPabloManifestFileName(fileName);
        }

        public static string NormalizeToLocalPath(string path)
        {
            if (path.StartsWith("file:", StringComparison.OrdinalIgnoreCase))
            {
                try
                {
                    return new Uri(path).LocalPath;
                }
                catch
                {
                    return path;
                }
            }

            return path;
        }

        private static string? TryGetFromDte()
        {
            try
            {
                if (ServiceProvider.GlobalProvider.GetService(typeof(DTE)) is not DTE dte)
                {
                    return null;
                }

                var document = dte.ActiveDocument;
                if (document == null)
                {
                    return null;
                }

                return NormalizeIfPabloManifest(document.FullName);
            }
            catch
            {
                return null;
            }
        }

        private static string? TryGetFromDocumentFrame()
        {
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
            return NormalizeIfPabloManifest(documentPathObj as string);
        }

        private static string? TryGetFromOpenDocuments()
        {
            try
            {
                if (ServiceProvider.GlobalProvider.GetService(typeof(DTE)) is not DTE dte)
                {
                    return null;
                }

                foreach (Document document in dte.Documents)
                {
                    var path = NormalizeIfPabloManifest(document.FullName);
                    if (path != null)
                    {
                        return path;
                    }
                }
            }
            catch
            {
                return null;
            }

            return null;
        }
    }
}
