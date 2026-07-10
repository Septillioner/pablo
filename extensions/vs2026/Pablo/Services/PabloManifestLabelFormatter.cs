using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;

namespace Pablo.VisualStudio.Services
{
    internal static class PabloManifestLabelFormatter
    {
        private const int MaxParentSegments = 6;

        public static IReadOnlyList<string> BuildUniqueLabels(IReadOnlyList<string> fullPaths)
        {
            ThreadHelper.ThrowIfNotOnUIThread();

            if (fullPaths.Count == 0)
            {
                return Array.Empty<string>();
            }

            var rootDirectory = TryGetSolutionDirectory();
            var labels = fullPaths
                .Select(path => FormatLabel(path, rootDirectory, parentSegments: 1))
                .ToList();

            for (var depth = 2; depth <= MaxParentSegments && HasDuplicates(labels); depth++)
            {
                for (var index = 0; index < fullPaths.Count; index++)
                {
                    labels[index] = FormatLabel(fullPaths[index], rootDirectory, depth);
                }
            }

            if (HasDuplicates(labels))
            {
                for (var index = 0; index < fullPaths.Count; index++)
                {
                    labels[index] = fullPaths[index];
                }
            }

            return labels;
        }

        private static bool HasDuplicates(IReadOnlyList<string> labels)
        {
            return labels.Distinct(StringComparer.OrdinalIgnoreCase).Count() < labels.Count;
        }

        private static string FormatLabel(string fullPath, string? rootDirectory, int parentSegments)
        {
            var normalizedPath = Path.GetFullPath(fullPath);
            var fileName = Path.GetFileName(normalizedPath);

            if (!string.IsNullOrWhiteSpace(rootDirectory))
            {
                var root = Path.GetFullPath(rootDirectory).TrimEnd(Path.DirectorySeparatorChar, Path.AltDirectorySeparatorChar);
                if (normalizedPath.StartsWith(root + Path.DirectorySeparatorChar, StringComparison.OrdinalIgnoreCase)
                    || normalizedPath.StartsWith(root + Path.AltDirectorySeparatorChar, StringComparison.OrdinalIgnoreCase))
                {
                    return normalizedPath.Substring(root.Length + 1);
                }
            }

            var directory = Path.GetDirectoryName(normalizedPath);
            if (string.IsNullOrWhiteSpace(directory))
            {
                return fileName;
            }

            var segments = new List<string> { fileName };
            var currentDirectory = directory;
            for (var segmentIndex = 0; segmentIndex < parentSegments && !string.IsNullOrWhiteSpace(currentDirectory); segmentIndex++)
            {
                segments.Insert(0, Path.GetFileName(currentDirectory));
                currentDirectory = Path.GetDirectoryName(currentDirectory);
            }

            return string.Join(Path.DirectorySeparatorChar.ToString(), segments);
        }

        private static string? TryGetSolutionDirectory()
        {
            try
            {
                var solution = ServiceProvider.GlobalProvider.GetService(typeof(SVsSolution)) as IVsSolution;
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
    }
}
