using System;
using System.Runtime.InteropServices;
using System.Threading.Tasks;
using Microsoft.VisualStudio.Shell;

namespace Pablo.VisualStudio.ToolWindows
{
    [Guid(PabloGuids.DeployToolWindowString)]
    public sealed class PabloDeployToolWindow : ToolWindowPane
    {
        public PabloDeployToolWindow() : base(provider: null)
        {
            Caption = "Pablo Run Deployment";
            Content = new PabloDeployControl();
        }

        public PabloDeployControl DeployControl => (PabloDeployControl)Content;

        public async Task RefreshAsync()
        {
            await DeployControl.RefreshAsync();
        }
    }
}
