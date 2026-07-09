using System;
using System.Collections.Generic;
using System.ComponentModel.Composition;
using System.Linq;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;
using System.Windows.Shapes;
using Microsoft.VisualStudio.Text;
using Microsoft.VisualStudio.Text.Editor;
using Microsoft.VisualStudio.Text.Formatting;
using Microsoft.VisualStudio.Text.Tagging;
using Microsoft.VisualStudio.Utilities;
using Pablo.VisualStudio.ContentType;

namespace Pablo.VisualStudio.Decorations
{
    internal sealed class ProfileStripeTag : ITag
    {
        public ProfileStripeTag(Color color, int indent, StripeKind kind)
        {
            Color = color;
            Indent = indent;
            Kind = kind;
        }

        public Color Color { get; }
        public int Indent { get; }
        public StripeKind Kind { get; }
    }

    [Export(typeof(AdornmentLayerDefinition))]
    [Name("PabloProfileStripe")]
    [Order(After = PredefinedAdornmentLayers.Selection, Before = PredefinedAdornmentLayers.Text)]
    internal sealed class ProfileStripeAdornmentLayerDefinition
    {
    }

    [Export(typeof(IViewTaggerProvider))]
    [ContentType(PabloContentTypeDefinitions.ContentType)]
    [TagType(typeof(ProfileStripeTag))]
    internal sealed class ProfileStripeTaggerProvider : IViewTaggerProvider
    {
        public ITagger<T>? CreateTagger<T>(ITextView textView, ITextBuffer buffer) where T : ITag
        {
            if (!typeof(T).IsAssignableFrom(typeof(ProfileStripeTag)))
            {
                return null;
            }

            return new ProfileStripeTagger(textView, buffer) as ITagger<T>;
        }
    }

    internal sealed class ProfileStripeTagger : ITagger<ProfileStripeTag>
    {
        private static readonly string[] ProfileColors =
        {
            "#5C6BC0", "#26A69A", "#EF5350", "#AB47BC", "#42A5F5", "#66BB6A",
        };

        private static readonly string[] EnvColors =
        {
            "#E53935", "#1E88E5", "#43A047", "#FB8C00", "#8E24AA", "#00ACC1",
            "#FDD835", "#6D4C41", "#EC407A", "#3949AB", "#00897B", "#C0CA33",
        };

        private readonly ITextView _textView;
        private readonly ITextBuffer _buffer;

        public ProfileStripeTagger(ITextView textView, ITextBuffer buffer)
        {
            _textView = textView;
            _buffer = buffer;
            _buffer.Changed += (_, __) => TagsChanged?.Invoke(this, new SnapshotSpanEventArgs(new SnapshotSpan(_buffer.CurrentSnapshot, 0, _buffer.CurrentSnapshot.Length)));
        }

        public event EventHandler<SnapshotSpanEventArgs>? TagsChanged;

        private string GetDocumentPath()
        {
            if (_buffer.Properties.TryGetProperty(typeof(ITextDocument), out ITextDocument? document) && document != null)
            {
                return document.FilePath ?? string.Empty;
            }

            return string.Empty;
        }

        public IEnumerable<ITagSpan<ProfileStripeTag>> GetTags(NormalizedSnapshotSpanCollection spans)
        {
            if (!PabloFileToContentTypeProvider.IsPabloManifestFileName(System.IO.Path.GetFileName(GetDocumentPath())))
            {
                yield break;
            }

            var marked = ProfileStripeParser.Parse(_buffer.CurrentSnapshot.GetText());
            var profileKeys = marked.Where(m => m.Kind == StripeKind.Profile).Select(m => m.ColorKey).Distinct().ToList();
            var envKeys = marked.Where(m => m.Kind == StripeKind.Env).Select(m => m.ColorKey).Distinct().ToList();
            var palette = BuildPalette(profileKeys, envKeys);

            foreach (var span in spans)
            {
                foreach (var entry in marked)
                {
                    if (entry.Line < span.Start.GetContainingLine().LineNumber || entry.Line > span.End.GetContainingLine().LineNumber)
                    {
                        continue;
                    }

                    if (!palette.TryGetValue(entry.ColorKey, out var color))
                    {
                        continue;
                    }

                    var line = _buffer.CurrentSnapshot.GetLineFromLineNumber(entry.Line);
                    var tagSpan = new SnapshotSpan(line.Start, line.End);
                    yield return new TagSpan<ProfileStripeTag>(tagSpan, new ProfileStripeTag(color, entry.Indent, entry.Kind));
                }
            }
        }

        private static Dictionary<string, Color> BuildPalette(IReadOnlyList<string> profileKeys, IReadOnlyList<string> envKeys)
        {
            var palette = new Dictionary<string, Color>(StringComparer.Ordinal);
            for (var i = 0; i < profileKeys.Count; i++)
            {
                palette[profileKeys[i]] = (Color)ColorConverter.ConvertFromString(ProfileColors[i % ProfileColors.Length]);
            }

            for (var i = 0; i < envKeys.Count; i++)
            {
                palette[envKeys[i]] = (Color)ColorConverter.ConvertFromString(EnvColors[i % EnvColors.Length]);
            }

            return palette;
        }
    }

    [Export(typeof(IWpfTextViewCreationListener))]
    [ContentType(PabloContentTypeDefinitions.ContentType)]
    [TextViewRole(PredefinedTextViewRoles.Document)]
    internal sealed class ProfileStripeAdornmentProvider : IWpfTextViewCreationListener
    {
        public void TextViewCreated(IWpfTextView textView)
        {
            var layer = textView.GetAdornmentLayer("PabloProfileStripe");
            if (layer == null)
            {
                return;
            }

            textView.LayoutChanged += (_, e) =>
            {
                if (!e.NewOrReformattedLines.Any())
                {
                    return;
                }

                layer.RemoveAllAdornments();
                var tagAggregator = AggregatorFactory.CreateTagAggregator<ProfileStripeTag>(textView);
                foreach (var line in e.NewOrReformattedLines)
                {
                    var lineSpan = line.Extent;
                    var tags = tagAggregator.GetTags(lineSpan);
                    foreach (var tagSpan in tags)
                    {
                        foreach (var snapshotSpan in tagSpan.Span.GetSpans(textView.TextSnapshot))
                        {
                            var geometry = textView.TextViewLines.GetTextMarkerGeometry(snapshotSpan);
                            if (geometry == null)
                            {
                                continue;
                            }

                            var tag = tagSpan.Tag;
                            ITextViewLine? containingLine = null;
                            foreach (var viewLine in textView.TextViewLines)
                            {
                                if (viewLine.Start <= snapshotSpan.Start && snapshotSpan.Start < viewLine.End)
                                {
                                    containingLine = viewLine;
                                    break;
                                }
                            }

                            if (containingLine == null)
                            {
                                continue;
                            }

                            var left = Math.Max(0, containingLine.Left + (tag.Indent * 4));
                            var rect = new Rectangle
                            {
                                Width = PabloConstants.DecorationBorderWidthPx,
                                Height = geometry.Bounds.Height,
                                Fill = new SolidColorBrush(tag.Color),
                            };

                            Canvas.SetLeft(rect, left);
                            Canvas.SetTop(rect, geometry.Bounds.Top);
                            layer.AddAdornment(AdornmentPositioningBehavior.TextRelative, snapshotSpan, tag, rect, null);
                        }
                    }
                }
            };
        }

        [Import]
        public IViewTagAggregatorFactoryService AggregatorFactory { get; set; } = null!;
    }
}
