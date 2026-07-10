using System;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using Microsoft.VisualStudio.PlatformUI;
using Microsoft.VisualStudio.Shell;
using Pablo.VisualStudio.Services;

namespace Pablo.VisualStudio.ToolWindows
{
    public partial class PabloDeployControl : UserControl
    {
        private const string NoBinaryMessage = "Pablo executable not set. Use Tools → Pablo: Select Executable.";
        private const string NoManifestMessage = "Open a pablo.yaml file in the editor, then click Refresh.";

        private InspectResult? _inspectResult;
        private string? _manifestPath;

        public PabloDeployControl()
        {
            InitializeComponent();
            Loaded += OnLoaded;
            Unloaded += OnUnloaded;
            ProfileCombo.SelectionChanged += (_, __) => UpdateEnvironmentCombo();
            RunButton.Click += (_, __) => _ = RunDeploymentAsync();
            RefreshButton.Click += (_, __) => _ = RefreshAsync();
        }

        private void OnLoaded(object sender, RoutedEventArgs e)
        {
            ApplyThemeBrushes();
            VSColorTheme.ThemeChanged += OnThemeChanged;
            _ = RefreshAsync();
        }

        private void OnUnloaded(object sender, RoutedEventArgs e)
        {
            VSColorTheme.ThemeChanged -= OnThemeChanged;
        }

        private void OnThemeChanged(ThemeChangedEventArgs e)
        {
            ApplyThemeBrushes();
        }

        private void ApplyThemeBrushes()
        {
            SetResourceReference(BackgroundProperty, EnvironmentColors.ToolWindowBackgroundBrushKey);
            RootGrid.SetResourceReference(Panel.BackgroundProperty, EnvironmentColors.ToolWindowBackgroundBrushKey);

            ApplyTextTheme(ManifestLabel);
            ApplyTextTheme(ManifestText);
            ApplyTextTheme(ProfileLabel);
            ApplyTextTheme(EnvironmentLabel);
            StatusText.SetResourceReference(TextBlock.ForegroundProperty, VsBrushes.GrayTextKey);

            ApplyThemedControlStyle(ProfileCombo, VsResourceKeys.ThemedDialogComboBoxStyleKey);
            ApplyThemedControlStyle(EnvironmentCombo, VsResourceKeys.ThemedDialogComboBoxStyleKey);
            ApplyThemedControlStyle(RunButton, VsResourceKeys.ThemedDialogButtonStyleKey);
            ApplyThemedControlStyle(RefreshButton, VsResourceKeys.ThemedDialogButtonStyleKey);
        }

        private static void ApplyTextTheme(TextBlock textBlock)
        {
            textBlock.SetResourceReference(TextBlock.ForegroundProperty, EnvironmentColors.ToolWindowTextBrushKey);
        }

        private static void ApplyThemedControlStyle(FrameworkElement element, object styleKey)
        {
            element.SetResourceReference(FrameworkElement.StyleProperty, styleKey);
        }

        public async Task RefreshAsync()
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();

            _manifestPath = PabloToolbarController.Instance.SelectedManifestPath
                ?? PabloManifestPathResolver.ResolveActiveManifestPath();
            if (_manifestPath == null)
            {
                SetUnavailableState(NoManifestMessage);
                return;
            }

            ManifestText.Text = _manifestPath;
            StatusText.Text = "Loading profiles...";
            RunButton.IsEnabled = false;

            var package = PabloPackage.Instance;
            var outcome = await package.InspectService.InspectManifestAsync(_manifestPath);
            switch (outcome.Status)
            {
                case InspectManifestStatus.NoBinary:
                    SetUnavailableState(NoBinaryMessage);
                    return;

                case InspectManifestStatus.Failed:
                    SetUnavailableState($"Inspect failed: {outcome.ErrorMessage}");
                    return;

                case InspectManifestStatus.NoProfiles:
                    SetUnavailableState("No profiles found in manifest.");
                    return;

                case InspectManifestStatus.Success:
                    _inspectResult = outcome.Result;
                    break;

                default:
                    SetUnavailableState("Failed to inspect manifest.");
                    return;
            }

            if (_inspectResult == null || _inspectResult.Profiles.Length == 0)
            {
                SetUnavailableState("No profiles found in manifest.");
                return;
            }

            StatusText.Text = $"{_inspectResult.Name} v{_inspectResult.Version}";
            ProfileCombo.ItemsSource = _inspectResult.Profiles;
            ProfileCombo.SelectedIndex = 0;
            UpdateEnvironmentCombo();
        }

        private void SetUnavailableState(string message)
        {
            StatusText.Text = message;
            ManifestText.Text = string.Empty;
            ProfileCombo.ItemsSource = null;
            EnvironmentCombo.ItemsSource = null;
            RunButton.IsEnabled = false;
        }

        private void UpdateEnvironmentCombo()
        {
            if (ProfileCombo.SelectedItem is not InspectProfile profile)
            {
                EnvironmentCombo.ItemsSource = null;
                RunButton.IsEnabled = false;
                return;
            }

            EnvironmentCombo.ItemsSource = profile.Environments;
            EnvironmentCombo.SelectedIndex = profile.Environments.Length > 0 ? 0 : -1;
            RunButton.IsEnabled = profile.Environments.Length > 0;
        }

        private async Task RunDeploymentAsync()
        {
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();

            if (string.IsNullOrWhiteSpace(_manifestPath))
            {
                StatusText.Text = "No manifest selected.";
                return;
            }

            if (ProfileCombo.SelectedItem is not InspectProfile profile)
            {
                StatusText.Text = "Select a profile.";
                return;
            }

            if (EnvironmentCombo.SelectedItem is not string environment)
            {
                StatusText.Text = "Select an environment.";
                return;
            }

            var runTarget = $"{profile.Name}/{environment}";
            var started = await PabloTerminalRunner.RunDeploymentAsync(
                PabloPackage.Instance.ExecutableService,
                _manifestPath,
                runTarget);

            StatusText.Text = started
                ? $"Started: pablo run -f {_manifestPath} {runTarget}"
                : NoBinaryMessage;
        }
    }
}
