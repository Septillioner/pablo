using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Text;

namespace Pablo.VisualStudio.Services
{
    internal static class PabloProcessHelper
    {
        // ProcessStartInfo.Arguments is CreateProcess argv — not a shell. Do not use PabloShell quoting.
        public static void SetArguments(ProcessStartInfo startInfo, IEnumerable<string> args)
        {
            startInfo.Arguments = string.Join(" ", args.Select(QuoteProcessArgument));
        }

        private static string QuoteProcessArgument(string arg)
        {
            if (string.IsNullOrEmpty(arg))
            {
                return "\"\"";
            }

            if (arg.IndexOfAny(new[] { ' ', '\t', '"' }) < 0)
            {
                return arg;
            }

            var builder = new StringBuilder(arg.Length + 2);
            builder.Append('"');
            builder.Append(arg.Replace("\"", "\\\""));
            builder.Append('"');
            return builder.ToString();
        }
    }
}
