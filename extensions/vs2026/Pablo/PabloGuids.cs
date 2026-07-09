using System;

namespace Pablo.VisualStudio
{
    internal static class PabloGuids
    {
        public const string PackageGuidString = "8f4e2a1b-3c5d-4e6f-9a0b-1c2d3e4f5a6b";
        public static readonly Guid PackageGuid = new(PackageGuidString);

        public const string CommandSetGuidString = "9a0b1c2d-3e4f-5a6b-7c8d-9e0f1a2b3c4d";
        public static readonly Guid CommandSetGuid = new(CommandSetGuidString);
    }

    internal enum PabloCommandIds : uint
    {
        Check = 0x0100,
        Init = 0x0101,
        Run = 0x0102,
        SelectExecutable = 0x0103,
        RunWithArgs = 0x0104,
    }
}
