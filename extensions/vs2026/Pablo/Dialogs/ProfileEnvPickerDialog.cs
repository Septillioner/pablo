using System.Linq;
using System.Windows;
using Pablo.VisualStudio.Services;

namespace Pablo.VisualStudio.Dialogs
{
    internal static class ProfileEnvPickerDialog
    {
        public static InspectProfile? PickProfile(InspectProfile[] profiles)
        {
            var window = new Window
            {
                Title = "Select profile",
                Width = 480,
                Height = 320,
                WindowStartupLocation = WindowStartupLocation.CenterScreen,
            };

            var list = new System.Windows.Controls.ListBox
            {
                DisplayMemberPath = "Name",
                ItemsSource = profiles,
            };

            window.Content = list;
            InspectProfile? selected = null;
            list.MouseDoubleClick += (_, __) =>
            {
                selected = list.SelectedItem as InspectProfile;
                window.DialogResult = true;
            };

            return window.ShowDialog() == true ? selected : null;
        }

        public static string? PickEnvironment(string[] environments)
        {
            var window = new Window
            {
                Title = "Select environment",
                Width = 360,
                Height = 280,
                WindowStartupLocation = WindowStartupLocation.CenterScreen,
            };

            var list = new System.Windows.Controls.ListBox
            {
                ItemsSource = environments,
            };

            window.Content = list;
            string? selected = null;
            list.MouseDoubleClick += (_, __) =>
            {
                selected = list.SelectedItem as string;
                window.DialogResult = true;
            };

            return window.ShowDialog() == true ? selected : null;
        }
    }
}
