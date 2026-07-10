using System.ComponentModel.Composition;
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
