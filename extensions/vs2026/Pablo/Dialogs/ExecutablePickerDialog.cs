using System.Collections.Generic;
using System.Linq;
using System.Windows;
using System.Windows.Controls;
using Microsoft.Win32;

namespace Pablo.VisualStudio.Dialogs
{
    internal static class ExecutablePickerDialog
    {
        public static string? PickExecutable(IReadOnlyList<string> candidates, string? current)
        {
            var window = new Window
            {
                Title = "Pablo: Select Executable",
                Width = 640,
                Height = 360,
                WindowStartupLocation = WindowStartupLocation.CenterScreen,
            };

            var list = new ListBox
            {
                ItemsSource = candidates,
            };

            if (!string.IsNullOrWhiteSpace(current))
            {
                var match = candidates.FirstOrDefault(c => c.Equals(current, System.StringComparison.OrdinalIgnoreCase));
                if (match != null)
                {
                    list.SelectedItem = match;
                }
            }

            var browseButton = new Button
            {
                Content = "Browse for executable...",
                Margin = new Thickness(8),
            };

            var panel = new DockPanel();
            DockPanel.SetDock(browseButton, Dock.Bottom);
            panel.Children.Add(browseButton);
            panel.Children.Add(list);
            window.Content = panel;

            string? selected = null;
            list.MouseDoubleClick += (_, __) =>
            {
                selected = list.SelectedItem as string;
                window.DialogResult = true;
            };

            browseButton.Click += (_, __) =>
            {
                var dialog = new OpenFileDialog
                {
                    Title = "Select Pablo executable",
                    Filter = "Executables (*.exe)|*.exe|All files (*.*)|*.*",
                };

                if (dialog.ShowDialog() == true)
                {
                    selected = dialog.FileName;
                    window.DialogResult = true;
                }
            };

            return window.ShowDialog() == true ? selected : null;
        }
    }
}
