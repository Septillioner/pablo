using System;
using System.Diagnostics;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;
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

        public async Task<InspectManifestOutcome> InspectManifestAsync(string filePath)
        {
            var uri = new Uri(filePath).AbsoluteUri;
            var lspResult = await PabloLanguageClientHost.TryListProfilesAsync(uri);
            if (lspResult?.Profiles != null && lspResult.Profiles.Length > 0)
            {
                return InspectManifestOutcome.Success(lspResult);
            }

            var binary = await _executableService.ResolveBinaryAsync();
            if (binary == null)
            {
                return InspectManifestOutcome.NoBinary();
            }

            try
            {
                var cliResult = await InspectViaCliAsync(binary, filePath);
                if (cliResult.Profiles.Length == 0)
                {
                    return InspectManifestOutcome.NoProfiles();
                }

                return InspectManifestOutcome.Success(cliResult);
            }
            catch (Exception ex)
            {
                _executableService.Log($"CLI inspect failed: {ex.Message}");
                return InspectManifestOutcome.Failed(ex.Message);
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
                var stderr = process.StandardError.ReadToEnd();
                process.WaitForExit();
                if (process.ExitCode != 0)
                {
                    var detail = string.IsNullOrWhiteSpace(stderr) ? $"exit code {process.ExitCode}" : stderr.Trim();
                    throw new InvalidOperationException($"pablo inspect failed: {detail}");
                }

                return JsonConvert.DeserializeObject<InspectResult>(stdout)
                    ?? throw new InvalidOperationException("Invalid inspect JSON output.");
            });
        }
    }
}
