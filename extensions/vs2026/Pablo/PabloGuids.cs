using System;

namespace Pablo.VisualStudio
{
    internal static class PabloGuids
    {
        public const string PackageGuidString = "8f4e2a1b-3c5d-4e6f-9a0b-1c2d3e4f5a6b";
        public static readonly Guid PackageGuid = new(PackageGuidString);

        public const string CommandSetGuidString = "9a0b1c2d-3e4f-5a6b-7c8d-9e0f1a2b3c4d";
        public static readonly Guid CommandSetGuid = new(CommandSetGuidString);

        public const string DeployToolWindowString = "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d";
        public static readonly Guid DeployToolWindowGuid = new(DeployToolWindowString);
    }

    internal enum PabloCommandIds : uint
    {
        Check = 0x0100,
        Init = 0x0101,
        Run = 0x0102,
        SelectExecutable = 0x0103,
        RunWithArgs = 0x0104,
        ShowDeploy = 0x0105,
        YamlCombo = 0x0106,
        YamlComboGetList = 0x0107,
        ProfileCombo = 0x0108,
        ProfileComboGetList = 0x0109,
        EnvCombo = 0x010a,
        EnvComboGetList = 0x010b,
        ToolbarRun = 0x010c,
    }
}
