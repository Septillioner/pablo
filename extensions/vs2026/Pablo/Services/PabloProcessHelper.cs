using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;

namespace Pablo.VisualStudio.Services
{
    internal static class PabloProcessHelper
    {
        public static void SetArguments(ProcessStartInfo startInfo, IEnumerable<string> args)
        {
            startInfo.Arguments = string.Join(" ", args.Select(arg => PabloShell.FormatArgForShell(arg)));
        }
    }
}
