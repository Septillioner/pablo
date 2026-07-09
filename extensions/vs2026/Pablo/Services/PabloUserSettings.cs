using System;
using System.IO;
using System.Text.Json;

namespace Pablo.VisualStudio.Services
{
    internal sealed class PabloUserSettings
    {
        private static readonly string SettingsDirectory = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
            "PabloVisualStudio");

        private static readonly string SettingsPath = Path.Combine(SettingsDirectory, "settings.json");

        public string? SelectedExecutable { get; set; }

        public static PabloUserSettings Load()
        {
            try
            {
                if (!File.Exists(SettingsPath))
                {
                    return new PabloUserSettings();
                }

                var json = File.ReadAllText(SettingsPath);
                return JsonSerializer.Deserialize<PabloUserSettings>(json) ?? new PabloUserSettings();
            }
            catch
            {
                return new PabloUserSettings();
            }
        }

        public void Save()
        {
            Directory.CreateDirectory(SettingsDirectory);
            var json = JsonSerializer.Serialize(this, new JsonSerializerOptions { WriteIndented = true });
            File.WriteAllText(SettingsPath, json);
        }
    }
}
