using System;
using System.Collections.Generic;
using System.Linq;

namespace Pablo.VisualStudio.Services
{
    internal enum ShellKind
    {
        PowerShell,
        Cmd,
        Posix,
    }

    internal static class PabloShell
    {
        private static readonly HashSet<string> PabloSubcommands = new(StringComparer.OrdinalIgnoreCase)
        {
            "run", "check", "init", "inspect", "uninstall", "version", "lsp",
        };

        public static ShellKind DetectShellKind()
        {
            var comspec = Environment.GetEnvironmentVariable("ComSpec") ?? string.Empty;
            if (comspec.Contains("powershell", StringComparison.OrdinalIgnoreCase) ||
                comspec.Contains("pwsh", StringComparison.OrdinalIgnoreCase))
            {
                return ShellKind.PowerShell;
            }

            if (comspec.EndsWith("cmd.exe", StringComparison.OrdinalIgnoreCase))
            {
                return ShellKind.Cmd;
            }

            return ShellKind.PowerShell;
        }

        public static string QuoteForShell(string value, ShellKind? kind = null)
        {
            kind ??= DetectShellKind();
            return kind switch
            {
                ShellKind.PowerShell => $"'{value.Replace("'", "''")}'",
                ShellKind.Cmd => $"\"{value.Replace("\"", "\\\"")}\"",
                _ => $"'{value.Replace("'", "'\\''")}'",
            };
        }

        private static bool ShouldQuoteArg(string arg)
        {
            if (arg.StartsWith("-", StringComparison.Ordinal))
            {
                return false;
            }

            return !PabloSubcommands.Contains(arg);
        }

        public static string FormatArgForShell(string arg, ShellKind? kind = null)
        {
            kind ??= DetectShellKind();
            return ShouldQuoteArg(arg) ? QuoteForShell(arg, kind) : arg;
        }

        public static string BuildTerminalCommand(string executable, IReadOnlyList<string> args, ShellKind? kind = null)
        {
            kind ??= DetectShellKind();
            var quotedExecutable = QuoteForShell(executable, kind);
            var formattedArgs = string.Join(" ", args.Select(arg => FormatArgForShell(arg, kind)));

            return kind == ShellKind.PowerShell
                ? $"& {quotedExecutable} {formattedArgs}"
                : $"{quotedExecutable} {formattedArgs}";
        }
    }
}
