using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;

namespace Pablo.VisualStudio.Services
{
    internal sealed class PabloToolbarController
    {
        public static PabloToolbarController Instance { get; } = new();

        private readonly object _gate = new();
        private IReadOnlyList<string> _manifestPaths = Array.Empty<string>();
        private InspectResult? _inspectResult;
        private int _selectedManifestIndex = -1;
        private int _selectedProfileIndex = -1;
        private int _selectedEnvironmentIndex = -1;
        private string? _loadedInspectManifestPath;

        public string? SelectedManifestPath
        {
            get
            {
                lock (_gate)
                {
                    return _selectedManifestIndex >= 0 && _selectedManifestIndex < _manifestPaths.Count
                        ? _manifestPaths[_selectedManifestIndex]
                        : null;
                }
            }
        }

        public string CurrentManifestLabel => GetManifestLabels().ElementAtOrDefault(GetSelectedManifestIndex()) ?? string.Empty;

        public string CurrentProfileLabel => GetProfileLabels().ElementAtOrDefault(GetSelectedProfileIndex()) ?? string.Empty;

        public string CurrentEnvironmentLabel => GetEnvironmentLabels().ElementAtOrDefault(GetSelectedEnvironmentIndex()) ?? string.Empty;

        public void RefreshManifests()
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            lock (_gate)
            {
                _manifestPaths = PabloManifestDiscovery.DiscoverManifestPaths();
                PabloOutputWindow.WriteLine($"Toolbar: discovered {_manifestPaths.Count} manifest(s).");

                if (_manifestPaths.Count == 0)
                {
                    _selectedManifestIndex = -1;
                    _inspectResult = null;
                    _loadedInspectManifestPath = null;
                    _selectedProfileIndex = -1;
                    _selectedEnvironmentIndex = -1;
                    return;
                }

                if (_selectedManifestIndex < 0 || _selectedManifestIndex >= _manifestPaths.Count)
                {
                    _selectedManifestIndex = 0;
                }
            }

            EnsureInspectForSelectedManifest();
        }

        public IReadOnlyList<string> GetManifestLabels()
        {
            lock (_gate)
            {
                return PabloManifestLabelFormatter.BuildUniqueLabels(_manifestPaths);
            }
        }

        public void SelectManifestByLabel(string label)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            var labels = GetManifestLabels();
            lock (_gate)
            {
                for (var index = 0; index < labels.Count; index++)
                {
                    if (!string.Equals(labels[index], label, StringComparison.OrdinalIgnoreCase))
                    {
                        continue;
                    }

                    if (index >= _manifestPaths.Count)
                    {
                        return;
                    }

                    _selectedManifestIndex = index;
                    _selectedProfileIndex = -1;
                    _selectedEnvironmentIndex = -1;
                    _inspectResult = null;
                    _loadedInspectManifestPath = null;
                    break;
                }
            }

            EnsureInspectForSelectedManifest();
        }

        public void SelectProfileByLabel(string label)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            EnsureInspectForSelectedManifest();
            lock (_gate)
            {
                if (_inspectResult?.Profiles == null)
                {
                    return;
                }

                var index = Array.FindIndex(_inspectResult.Profiles, profile =>
                    string.Equals(profile.Name, label, StringComparison.OrdinalIgnoreCase));
                if (index < 0)
                {
                    return;
                }

                _selectedProfileIndex = index;
                _selectedEnvironmentIndex = -1;
            }
        }

        public void SelectEnvironmentByLabel(string label)
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            EnsureInspectForSelectedManifest();
            lock (_gate)
            {
                var profile = GetSelectedProfileLocked();
                if (profile == null)
                {
                    return;
                }

                var index = Array.FindIndex(profile.Environments, env =>
                    string.Equals(env, label, StringComparison.OrdinalIgnoreCase));
                if (index < 0)
                {
                    return;
                }

                _selectedEnvironmentIndex = index;
            }
        }

        public IReadOnlyList<string> GetProfileLabels()
        {
            EnsureInspectForSelectedManifest();
            lock (_gate)
            {
                if (_inspectResult?.Profiles == null || _inspectResult.Profiles.Length == 0)
                {
                    return Array.Empty<string>();
                }

                if (_selectedProfileIndex < 0)
                {
                    _selectedProfileIndex = 0;
                }

                return _inspectResult.Profiles.Select(profile => profile.Name).ToList();
            }
        }

        public IReadOnlyList<string> GetEnvironmentLabels()
        {
            EnsureInspectForSelectedManifest();
            lock (_gate)
            {
                var profile = GetSelectedProfileLocked();
                if (profile == null || profile.Environments.Length == 0)
                {
                    return Array.Empty<string>();
                }

                if (_selectedEnvironmentIndex < 0)
                {
                    _selectedEnvironmentIndex = 0;
                }

                return profile.Environments.ToList();
            }
        }

        public void EnsureInspectForSelectedManifest()
        {
            ThreadHelper.ThrowIfNotOnUIThread();
            string? manifestPath;
            lock (_gate)
            {
                manifestPath = SelectedManifestPath;
                if (string.IsNullOrWhiteSpace(manifestPath))
                {
                    return;
                }

                if (string.Equals(manifestPath, _loadedInspectManifestPath, StringComparison.OrdinalIgnoreCase)
                    && _inspectResult != null)
                {
                    return;
                }
            }

            ThreadHelper.JoinableTaskFactory.Run(LoadInspectForSelectedManifestAsync);
        }

        public async Task<bool> RunSelectedDeploymentAsync()
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            EnsureInspectForSelectedManifest();

            string? manifestPath;
            string? runTarget;
            lock (_gate)
            {
                manifestPath = SelectedManifestPath;
                var profile = GetSelectedProfileLocked();
                if (profile == null || _selectedEnvironmentIndex < 0 || _selectedEnvironmentIndex >= profile.Environments.Length)
                {
                    return false;
                }

                runTarget = $"{profile.Name}/{profile.Environments[_selectedEnvironmentIndex]}";
            }

            if (string.IsNullOrWhiteSpace(manifestPath) || string.IsNullOrWhiteSpace(runTarget))
            {
                return false;
            }

            var executableService = PabloPackage.Instance?.ExecutableService;
            if (executableService == null)
            {
                return false;
            }

            return await PabloTerminalRunner.RunDeploymentAsync(executableService, manifestPath, runTarget);
        }

        private int GetSelectedManifestIndex()
        {
            lock (_gate)
            {
                return _selectedManifestIndex;
            }
        }

        private int GetSelectedProfileIndex()
        {
            lock (_gate)
            {
                return _selectedProfileIndex;
            }
        }

        private int GetSelectedEnvironmentIndex()
        {
            lock (_gate)
            {
                return _selectedEnvironmentIndex;
            }
        }

        private InspectProfile? GetSelectedProfileLocked()
        {
            if (_inspectResult?.Profiles == null || _inspectResult.Profiles.Length == 0)
            {
                return null;
            }

            if (_selectedProfileIndex < 0 || _selectedProfileIndex >= _inspectResult.Profiles.Length)
            {
                _selectedProfileIndex = 0;
            }

            return _inspectResult.Profiles[_selectedProfileIndex];
        }

        private async Task LoadInspectForSelectedManifestAsync()
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();

            string? manifestPath;
            lock (_gate)
            {
                manifestPath = SelectedManifestPath;
            }

            if (string.IsNullOrWhiteSpace(manifestPath))
            {
                return;
            }

            var inspectService = PabloPackage.Instance?.InspectService;
            if (inspectService == null)
            {
                return;
            }

            var outcome = await inspectService.InspectManifestAsync(manifestPath);
            lock (_gate)
            {
                _loadedInspectManifestPath = manifestPath;
                _inspectResult = outcome.Status == InspectManifestStatus.Success ? outcome.Result : null;
                _selectedProfileIndex = _inspectResult?.Profiles.Length > 0 ? 0 : -1;
                _selectedEnvironmentIndex = GetSelectedProfileLocked()?.Environments.Length > 0 ? 0 : -1;
            }
        }
    }
}
