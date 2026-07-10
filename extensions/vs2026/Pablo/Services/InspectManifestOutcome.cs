namespace Pablo.VisualStudio.Services
{
    public enum InspectManifestStatus
    {
        Success,
        NoBinary,
        Failed,
        NoProfiles,
    }

    public sealed class InspectManifestOutcome
    {
        private InspectManifestOutcome(InspectManifestStatus status, InspectResult? result, string? errorMessage)
        {
            Status = status;
            Result = result;
            ErrorMessage = errorMessage;
        }

        public InspectManifestStatus Status { get; }

        public InspectResult? Result { get; }

        public string? ErrorMessage { get; }

        public static InspectManifestOutcome Success(InspectResult result) =>
            new(InspectManifestStatus.Success, result, null);

        public static InspectManifestOutcome NoBinary() =>
            new(InspectManifestStatus.NoBinary, null, null);

        public static InspectManifestOutcome Failed(string message) =>
            new(InspectManifestStatus.Failed, null, message);

        public static InspectManifestOutcome NoProfiles() =>
            new(InspectManifestStatus.NoProfiles, null, null);
    }
}
