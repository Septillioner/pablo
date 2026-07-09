using System;
using System.Threading.Tasks;
using Microsoft.VisualStudio.LanguageServer.Client;
using Microsoft.VisualStudio.Shell;
using Newtonsoft.Json.Linq;

namespace Pablo.VisualStudio.Lsp
{
    internal sealed class PabloMiddleLayer : ILanguageClientMiddleLayer
    {
        public bool CanHandle(string methodName)
        {
            return methodName == "textDocument/codeLens" || methodName == "workspace/executeCommand";
        }

        public async Task<JToken?> HandleRequestAsync(
            string methodName,
            JToken methodParam,
            Func<JToken, Task<JToken?>> sendRequest)
        {
            if (methodName == "workspace/executeCommand")
            {
                return await HandleExecuteCommandAsync(methodParam).ConfigureAwait(false);
            }

            var response = await sendRequest(methodParam).ConfigureAwait(false);
            if (response is not JArray lenses)
            {
                return response;
            }

            foreach (var lens in lenses)
            {
                if (lens?["command"]?["command"]?.Value<string>() != PabloConstants.RunWithArgsCommand)
                {
                    continue;
                }

                lens["command"]!["title"] = "Run";
                var runTarget = lens["command"]?["arguments"]?[1]?.Value<string>();
                if (!string.IsNullOrWhiteSpace(runTarget))
                {
                    lens["command"]!["tooltip"] = runTarget;
                }
            }

            return lenses;
        }

        public Task HandleNotificationAsync(
            string methodName,
            JToken methodParam,
            Func<JToken, Task> sendNotification)
        {
            return sendNotification(methodParam);
        }

        private static async Task<JToken?> HandleExecuteCommandAsync(JToken methodParam)
        {
            if (methodParam["command"]?.Value<string>() != PabloConstants.RunWithArgsCommand)
            {
                return null;
            }

            if (methodParam["arguments"] is not JArray args || args.Count < 2)
            {
                return JValue.CreateNull();
            }

            var uri = args[0]?.Value<string>();
            var runTarget = args[1]?.Value<string>();
            if (string.IsNullOrWhiteSpace(uri) || string.IsNullOrWhiteSpace(runTarget))
            {
                return JValue.CreateNull();
            }

            var filePath = new Uri(uri).LocalPath;
            await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
            await PabloTerminalRunner.RunDeploymentAsync(PabloPackage.Instance.ExecutableService, filePath, runTarget);
            return JValue.CreateNull();
        }
    }
}
