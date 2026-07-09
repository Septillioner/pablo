using System;
using System.Collections.Generic;
using System.ComponentModel.Composition;
using System.Diagnostics;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.VisualStudio.LanguageServer.Client;
using Microsoft.VisualStudio.Threading;
using Microsoft.VisualStudio.Utilities;
using Pablo.VisualStudio.ContentType;
using Pablo.VisualStudio.Services;
using StreamJsonRpc;

namespace Pablo.VisualStudio.Lsp
{
    [Export(typeof(ILanguageClient))]
    [ContentType(PabloContentTypeDefinitions.ContentType)]
    internal sealed class PabloLanguageClient : ILanguageClient, ILanguageClientCustomMessage2
    {
        private readonly PabloMiddleLayer _middleLayer = new();

        public string Name => "Pablo Language Server";

        public IEnumerable<string>? ConfigurationSections => null;

        public object? InitializationOptions => null;

        public IEnumerable<string>? FilesToWatch => new[] { "**/pablo*.yaml", "**/pablo*.yml", "**/pablo.yaml", "**/pablo.yml" };

        public object? MiddleLayer => _middleLayer;

        public object? CustomMessageTarget => null;

        public bool ShowNotificationOnInitializeFailed => true;

        public JsonRpc? Rpc { get; private set; }

        public event AsyncEventHandler<EventArgs>? StartAsync;
        public event AsyncEventHandler<EventArgs>? StopAsync;

        public async Task<Connection?> ActivateAsync(CancellationToken token)
        {
            PabloLanguageClientHost.Register(this);
            var binary = await PabloPackage.Instance.ExecutableService.ResolveBinaryAsync();
            if (binary == null)
            {
                PabloOutputWindow.WriteLine("Pablo binary not found for LSP activation.");
                return null;
            }

            if (!await PabloPackage.Instance.ExecutableService.AssertLspSupportedAsync(binary))
            {
                PabloOutputWindow.WriteLine($"Pablo binary does not support LSP: {binary}");
                return null;
            }

            PabloOutputWindow.WriteLine($"Using Pablo binary: {binary}");

            var startInfo = new ProcessStartInfo
            {
                FileName = binary,
                Arguments = "lsp",
                RedirectStandardInput = true,
                RedirectStandardOutput = true,
                UseShellExecute = false,
                CreateNoWindow = true,
            };

            var process = Process.Start(startInfo);
            if (process == null)
            {
                return null;
            }

            return new Connection(process.StandardOutput.BaseStream, process.StandardInput.BaseStream);
        }

        public async Task OnLoadedAsync()
        {
            if (StartAsync != null)
            {
                await StartAsync.InvokeAsync(this, EventArgs.Empty);
            }
        }

        public Task OnServerInitializedAsync()
        {
            return Task.CompletedTask;
        }

        public Task<InitializationFailureContext?> OnServerInitializeFailedAsync(ILanguageClientInitializationInfo initializationState)
        {
            return Task.FromResult<InitializationFailureContext?>(new InitializationFailureContext
            {
                FailureMessage = initializationState?.StatusMessage ?? "Pablo language server failed to initialize.",
            });
        }

        public Task AttachForCustomMessageAsync(JsonRpc rpc)
        {
            Rpc = rpc;
            return Task.CompletedTask;
        }

        public Task DetachAsync()
        {
            Rpc = null;
            return Task.CompletedTask;
        }

        internal async Task RestartLanguageServerAsync()
        {
            if (StopAsync != null)
            {
                await StopAsync.InvokeAsync(this, EventArgs.Empty);
            }

            if (StartAsync != null)
            {
                await StartAsync.InvokeAsync(this, EventArgs.Empty);
            }
        }
    }
}
