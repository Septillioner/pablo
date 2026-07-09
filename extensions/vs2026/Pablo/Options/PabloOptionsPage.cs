using System.ComponentModel;
using Microsoft.VisualStudio.Shell;

namespace Pablo.VisualStudio.Options
{
    internal sealed class PabloOptionsPage : DialogPage
    {
        public static PabloOptionsPage? Instance { get; private set; }

        [Category("Pablo")]
        [DisplayName("Executable path")]
        [Description("Absolute path to the Pablo CLI binary. When empty, the extension uses a selected binary or pablo from PATH.")]
        public string ExecutablePath { get; set; } = string.Empty;

        public override void LoadSettingsFromStorage()
        {
            base.LoadSettingsFromStorage();
            Instance = this;
        }

        protected override void OnApply(PageApplyEventArgs e)
        {
            base.OnApply(e);
            PabloPackage.Instance?.OnOptionsChanged();
        }
    }
}
