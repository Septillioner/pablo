using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using EnvDTE;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;
using Pablo.VisualStudio.ContentType;

namespace Pablo.VisualStudio.Services
{
    internal static class PabloManifestDiscovery
    {
        private const int MaxSearchDepth = 4;

        private static readonly HashSet<string> SkippedDirectoryNames = new(StringComparer.OrdinalIgnoreCase)
        {
            ".git",
            ".vs",
            "node_modules",
            "bin",
            "obj",
            "packages",
        };

        public static IReadOnlyList<string> DiscoverManifestPaths()
        {
            ThreadHelper.ThrowIfNotOnUIThread();

            var results = new List<string>();
            var seen = new HashSet<string>(StringComparer.OrdinalIgnoreCase);

            try
            {
                if (ServiceProvider.GlobalProvider.GetService(typeof(DTE)) is DTE dte)
                {
                    if (dte.Solution != null && dte.Solution.IsOpen && dte.Solution.Projects != null)
                    {
                        foreach (Project project in dte.Solution.Projects)
                        {
                            CollectFromProject(project, results, seen);
                        }
                    }

                    if (dte.Documents != null)
                    {
                        foreach (Document document in dte.Documents)
                        {
                            AddIfManifest(document.FullName, results, seen);
                        }
                    }
                }

                foreach (var root in GetSearchRoots())
                {
                    CollectFromDirectory(root, results, seen, currentDepth: 0);
                }

                AddIfManifest(PabloManifestPathResolver.ResolveActiveManifestPath(), results, seen);
            }
            catch
            {
                // DTE enumeration can fail for unloaded projects.
            }

            return results.OrderBy(path => path, StringComparer.OrdinalIgnoreCase).ToList();
        }

        private static IEnumerable<string> GetSearchRoots()
        {
            var roots = new HashSet<string>(StringComparer.OrdinalIgnoreCase);

            try
            {
                var solution = ServiceProvider.GlobalProvider.GetService(typeof(SVsSolution)) as IVsSolution;
                if (solution != null)
                {
                    solution.GetSolutionInfo(out string? solutionDirectory, out _, out _);
                    if (!string.IsNullOrWhiteSpace(solutionDirectory) && Directory.Exists(solutionDirectory))
                    {
                        roots.Add(Path.GetFullPath(solutionDirectory));
                    }
                }
            }
            catch
            {
                // Solution info may be unavailable during early load.
            }

            return roots;
        }

        private static void CollectFromDirectory(string directory, List<string> results, HashSet<string> seen, int currentDepth)
        {
            if (currentDepth > MaxSearchDepth || !Directory.Exists(directory))
            {
                return;
            }

            try
            {
                foreach (var file in Directory.EnumerateFiles(directory))
                {
                    AddIfManifest(file, results, seen);
                }

                if (currentDepth >= MaxSearchDepth)
                {
                    return;
                }

                foreach (var childDirectory in Directory.EnumerateDirectories(directory))
                {
                    var directoryName = Path.GetFileName(childDirectory);
                    if (SkippedDirectoryNames.Contains(directoryName))
                    {
                        continue;
                    }

                    CollectFromDirectory(childDirectory, results, seen, currentDepth + 1);
                }
            }
            catch
            {
                // Skip directories that cannot be read.
            }
        }

        private static void CollectFromProject(Project project, List<string> results, HashSet<string> seen)
        {
            if (project == null)
            {
                return;
            }

            try
            {
                if (string.Equals(project.Kind, EnvDTE80.ProjectKinds.vsProjectKindSolutionFolder, StringComparison.OrdinalIgnoreCase))
                {
                    foreach (ProjectItem item in EnumerateProjectItems(project.ProjectItems))
                    {
                        if (item.SubProject != null)
                        {
                            CollectFromProject(item.SubProject, results, seen);
                        }
                        else
                        {
                            CollectFromProjectItem(item, results, seen);
                        }
                    }

                    return;
                }

                foreach (ProjectItem item in EnumerateProjectItems(project.ProjectItems))
                {
                    CollectFromProjectItem(item, results, seen);
                }
            }
            catch
            {
                // Skip projects that cannot enumerate items.
            }
        }

        private static void CollectFromProjectItem(ProjectItem item, List<string> results, HashSet<string> seen)
        {
            if (item == null)
            {
                return;
            }

            try
            {
                if (item.SubProject != null)
                {
                    CollectFromProject(item.SubProject, results, seen);
                    return;
                }

                if (item.FileCount > 0)
                {
                    for (var i = 1; i <= item.FileCount; i++)
                    {
                        AddIfManifest(item.FileNames[(short)i], results, seen);
                    }
                }

                foreach (ProjectItem child in EnumerateProjectItems(item.ProjectItems))
                {
                    CollectFromProjectItem(child, results, seen);
                }
            }
            catch
            {
                // Skip items that cannot be read.
            }
        }

        private static IEnumerable<ProjectItem> EnumerateProjectItems(ProjectItems? items)
        {
            if (items == null)
            {
                yield break;
            }

            foreach (ProjectItem item in items)
            {
                if (item != null)
                {
                    yield return item;
                }
            }
        }

        private static void AddIfManifest(string? path, List<string> results, HashSet<string> seen)
        {
            if (string.IsNullOrWhiteSpace(path))
            {
                return;
            }

            var localPath = PabloManifestPathResolver.NormalizeToLocalPath(path);
            if (!PabloFileToContentTypeProvider.IsPabloManifestFileName(Path.GetFileName(localPath)))
            {
                return;
            }

            if (!File.Exists(localPath))
            {
                return;
            }

            var fullPath = Path.GetFullPath(localPath);
            if (seen.Add(fullPath))
            {
                results.Add(fullPath);
            }
        }
    }
}
