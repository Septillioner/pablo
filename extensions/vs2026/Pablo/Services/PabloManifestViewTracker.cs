using System.ComponentModel.Composition;
using Microsoft.VisualStudio.Shell;
using Microsoft.VisualStudio.Text;
using Microsoft.VisualStudio.Text.Editor;
using Microsoft.VisualStudio.Utilities;
using Pablo.VisualStudio.ContentType;

namespace Pablo.VisualStudio.Services
{
    [Export(typeof(IWpfTextViewCreationListener))]
    [ContentType(PabloContentTypeDefinitions.ContentType)]
    [TextViewRole(PredefinedTextViewRoles.Document)]
    internal sealed class PabloManifestViewTracker : IWpfTextViewCreationListener
    {
        public void TextViewCreated(IWpfTextView textView)
        {
            TrackManifestPath(textView);
            textView.GotAggregateFocus += (_, __) => TrackManifestPath(textView);

            if (!textView.TextBuffer.Properties.TryGetProperty(typeof(ITextDocument), out ITextDocument? document)
                || document == null)
            {
                return;
            }

            void OnFileActionOccurred(object? sender, TextDocumentFileActionEventArgs e)
            {
                if (e.FileActionType != FileActionTypes.ContentSavedToDisk)
                {
                    return;
                }

                var filePath = document.FilePath;
                ThreadHelper.JoinableTaskFactory.RunAsync(async () =>
                {
                    await ThreadHelper.JoinableTaskFactory.SwitchToMainThreadAsync();
                    PabloToolbarController.Instance.NotifyManifestSaved(filePath);
                });
            }

            document.FileActionOccurred += OnFileActionOccurred;
            textView.Closed += (_, __) => document.FileActionOccurred -= OnFileActionOccurred;
        }

        private static void TrackManifestPath(IWpfTextView textView)
        {
            if (textView.TextBuffer.Properties.TryGetProperty(typeof(ITextDocument), out ITextDocument? document) && document != null)
            {
                PabloManifestTracker.SetLastKnown(document.FilePath);
            }
        }
    }
}
