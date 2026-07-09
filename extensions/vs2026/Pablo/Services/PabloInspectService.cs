using System;
using System.Diagnostics;
using System.IO;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Shell.Interop;
using Newtonsoft.Json;
using Pablo.VisualStudio.Lsp;

namespace Pablo.VisualStudio.Services
{
    public sealed class InspectProfile
    {
        [JsonProperty("name")]
        public string Name { get; set; } = string.Empty;

        [JsonProperty("type")]
        public string Type { get; set; } = string.Empty;

        [JsonProperty("environments")]
        public string[] Environments { get; set; } = Array.Empty<string>();
    }

    public sealed class InspectResult
    {
        [JsonProperty("name")]
        public string Name { get; set; } = string.Empty;

        [JsonProperty("version")]
        public string Version { get; set; } = string.Empty;

        [JsonProperty("profiles")]
        public InspectProfile[] Profiles { get; set; } = Array.Empty<InspectProfile>();
    }

    public sealed class PabloInspectService
    {
        private readonly PabloExecutableService _executableService;

        public PabloInspectService(PabloExecutableService executableService)
        {
            _executableService = executableService;
        }

        public async Task<InspectResult?> InspectManifestAsync(string filePath)
        {
            var uri = new Uri(filePath).AbsoluteUri;
            var lspResult = await PabloLanguageClientHost.TryListProfilesAsync(uri);
            if (lspResult?.Profiles != null)
            {
                return lspResult;
            }

            var binary = await _executableService.ResolveBinaryAsync();
            if (binary == null)
            {
                return null;
            }

            try
            {
                return await InspectViaCliAsync(binary, filePath);
            }
            catch (Exception ex)
            {
                _executableService.Log($"CLI inspect failed: {ex.Message}");
                VsShellUtilities.ShowMessageBox(
                    ServiceProvider.GlobalProvider,
                    $"Failed to inspect manifest: {ex.Message}",
                    "Pablo",
                    OLEMSGICON.OLEMSGICON_CRITICAL,
                    OLEMSGBUTTON.OLEMSGBUTTON_OK,
                    OLEMSGDEFBUTTON.OLEMSGDEFBUTTON_FIRST);
                return null;
            }
        }

        private static async Task<InspectResult> InspectViaCliAsync(string binary, string filePath)
        {
            return await Task.Run(() =>
            {
                var startInfo = new ProcessStartInfo
                {
                    FileName = binary,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                };
                PabloProcessHelper.SetArguments(startInfo, new[] { "inspect", "-f", filePath, "--json" });

                using var process = Process.Start(startInfo) ?? throw new InvalidOperationException("Failed to start pablo inspect.");
                var stdout = process.StandardOutput.ReadToEnd();
                process.WaitForExit();
                if (process.ExitCode != 0)
                {
                    throw new InvalidOperationException($"pablo inspect exited with code {process.ExitCode}");
                }

                return JsonConvert.DeserializeObject<InspectResult>(stdout)
                    ?? throw new InvalidOperationException("Invalid inspect JSON output.");
            });
        }
    }
}
