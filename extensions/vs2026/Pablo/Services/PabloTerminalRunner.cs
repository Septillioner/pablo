using System;
using System.Diagnostics;
using System.Threading.Tasks;
using Pablo.VisualStudio.Services;

namespace Pablo.VisualStudio
{
    internal static class PabloTerminalRunner
    {
        public static async Task<bool> RunCliAsync(PabloExecutableService executableService, string[] args)
        {
            var binary = await executableService.ResolveBinaryAsync();
            if (binary == null)
            {
                await PabloCommandHandler.PromptMissingExecutableAsync(executableService);
                return false;
            }

            // Always launch via cmd.exe with /s so multi-quoted paths are not mangled.
            // (cmd /k "exe" arg "file" strips quotes incorrectly → ERROR_INVALID_NAME.)
            var line = PabloShell.BuildTerminalCommand(binary, args, ShellKind.Cmd);
            Process.Start(new ProcessStartInfo
            {
                FileName = "cmd.exe",
                Arguments = "/s /k \"" + line + "\"",
                UseShellExecute = false,
            });
            return true;
        }

        public static async Task<bool> RunDeploymentAsync(PabloExecutableService executableService, string filePath, string runTarget)
        {
            return await RunCliAsync(executableService, new[] { "run", "-f", filePath, runTarget });
        }
    }
}
