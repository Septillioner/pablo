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
    [ContentType("yaml")]
    [ContentType("YAML")]
    internal sealed class PabloLanguageClient : ILanguageClient, ILanguageClientCustomMessage2
    {
        private readonly PabloMiddleLayer _middleLayer = new();
        private Process? _activeProcess;

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

            try
            {
                var package = PabloPackage.Instance;
                var binary = await PabloBinaryResolver.ResolveBinaryAsync(package?.ExecutableService);
                if (binary == null)
                {
                    PabloOutputWindow.WriteLine("Pablo binary not found for LSP activation. Use Tools → Pablo: Select Executable.");
                    return null;
                }

                if (!await PabloBinaryResolver.AssertLspSupportedAsync(binary))
                {
                    PabloOutputWindow.WriteLine($"Pablo binary does not support LSP: {binary}");
                    return null;
                }

                PabloOutputWindow.WriteLine($"Starting Pablo LSP: {binary}");

                var startInfo = new ProcessStartInfo
                {
                    FileName = binary,
                    RedirectStandardInput = true,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                };
                PabloProcessHelper.SetArguments(startInfo, new[] { "lsp" });

                var process = Process.Start(startInfo);
                if (process == null)
                {
                    PabloOutputWindow.WriteLine("Failed to start pablo lsp process.");
                    return null;
                }

                _activeProcess = process;
                PabloOutputWindow.WriteLine($"Pablo LSP started (pid {process.Id}).");

                _ = Task.Run(async () =>
                {
                    try
                    {
                        var stderr = await process.StandardError.ReadToEndAsync();
                        if (!string.IsNullOrWhiteSpace(stderr))
                        {
                            PabloOutputWindow.WriteLine($"pablo lsp stderr: {stderr.Trim()}");
                        }
                    }
                    catch
                    {
                        // Process may exit before stderr is fully read.
                    }
                }, token);

                _ = Task.Run(() =>
                {
                    try
                    {
                        process.WaitForExit();
                        PabloOutputWindow.WriteLine($"Pablo LSP exited (pid {process.Id}, code {process.ExitCode}).");
                    }
                    catch
                    {
                        // Process may already be gone during shutdown.
                    }
                }, token);

                return new Connection(process.StandardOutput.BaseStream, process.StandardInput.BaseStream);
            }
            catch (Exception ex)
            {
                PabloOutputWindow.WriteLine($"Pablo LSP activation failed: {ex.Message}");
                return null;
            }
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
            PabloOutputWindow.WriteLine("Pablo LSP initialized.");
            return Task.CompletedTask;
        }

        public Task<InitializationFailureContext?> OnServerInitializeFailedAsync(ILanguageClientInitializationInfo initializationState)
        {
            PabloOutputWindow.WriteLine($"Pablo LSP initialize failed: {initializationState?.StatusMessage ?? "unknown error"}");
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
            _activeProcess = null;
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
