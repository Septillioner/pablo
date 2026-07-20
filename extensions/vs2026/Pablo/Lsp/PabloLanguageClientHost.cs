using System;
using System.Threading;
using System.Threading.Tasks;
using Newtonsoft.Json.Linq;
using Pablo.VisualStudio.Services;

namespace Pablo.VisualStudio.Lsp
{
    internal static class PabloLanguageClientHost
    {
        private static PabloLanguageClient? _client;

        public static void Register(PabloLanguageClient client)
        {
            _client = client;
        }

        public static bool IsRunning => _client?.Rpc != null;

        public static async Task StopAsync()
        {
            if (_client == null)
            {
                return;
            }

            await _client.StopLanguageServerAsync();
        }

        public static async Task RestartAsync()
        {
            if (_client == null)
            {
                return;
            }

            await _client.RestartLanguageServerAsync();
        }

        public static async Task<InspectResult?> TryListProfilesAsync(string uri, CancellationToken cancellationToken = default)
        {
            var rpc = _client?.Rpc;
            if (rpc == null)
            {
                return null;
            }

            try
            {
                var result = await rpc.InvokeAsync<JObject>(
                    PabloConstants.ListProfilesMethod,
                    new { uri },
                    cancellationToken).ConfigureAwait(false);

                return result?.ToObject<InspectResult>();
            }
            catch (Exception ex)
            {
                PabloOutputWindow.WriteLine($"LSP listProfiles failed, falling back to CLI: {ex.Message}");
                return null;
            }
        }
    }
}
