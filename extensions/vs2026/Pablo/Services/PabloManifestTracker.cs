using System;

namespace Pablo.VisualStudio.Services
{
    internal static class PabloManifestTracker
    {
        private static string? _lastKnownPath;

        public static string? LastKnownPath => _lastKnownPath;

        public static void SetLastKnown(string? path)
        {
            if (string.IsNullOrWhiteSpace(path))
            {
                return;
            }

            var normalized = PabloManifestPathResolver.NormalizeToLocalPath(path);
            if (PabloManifestPathResolver.IsPabloManifestFileName(System.IO.Path.GetFileName(normalized)))
            {
                _lastKnownPath = normalized;
            }
        }
    }
}
