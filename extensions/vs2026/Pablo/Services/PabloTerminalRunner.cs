using System;
using System.Diagnostics;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;
using Pablo.VisualStudio.Services;

namespace Pablo.VisualStudio
{
    internal static class PabloTerminalRunner
    {
        public static async Task RunCliAsync(PabloExecutableService executableService, string[] args)
        {
            var binary = await executableService.ResolveBinaryAsync();
            if (binary == null)
            {
                await PabloCommandHandler.PromptMissingExecutableAsync(executableService);
                return;
            }

            var line = PabloShell.BuildTerminalCommand(binary, args);
            Process.Start(new ProcessStartInfo
            {
                FileName = "cmd.exe",
                Arguments = $"/k {line}",
                UseShellExecute = true,
            });
        }

        public static async Task RunDeploymentAsync(PabloExecutableService executableService, string filePath, string runTarget)
        {
            await RunCliAsync(executableService, new[] { "run", "-f", filePath, runTarget });
        }
    }
}
