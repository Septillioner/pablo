using System;
using System.ComponentModel.Composition;
using System.IO;
using Microsoft.VisualStudio.Utilities;

namespace Pablo.VisualStudio.ContentType
{
    internal abstract class PabloFilePathProviderBase : IFilePathToContentTypeProvider
    {
        [Import]
        public IContentTypeRegistryService ContentTypeRegistry { get; set; } = null!;

        public bool TryGetContentTypeForFilePath(string filePath, out IContentType? contentType)
        {
            contentType = null;
            if (string.IsNullOrWhiteSpace(filePath))
            {
                return false;
            }

            if (!PabloFileToContentTypeProvider.IsPabloManifestFileName(Path.GetFileName(filePath)))
            {
                return false;
            }

            contentType = ContentTypeRegistry.GetContentType(PabloContentTypeDefinitions.ContentType);
            return contentType != null;
        }
    }

    [Export(typeof(IFilePathToContentTypeProvider))]
    [Name("PabloYamlFilePathProvider")]
    [FileExtension(".yaml")]
    internal sealed class PabloYamlFilePathProvider : PabloFilePathProviderBase
    {
    }

    [Export(typeof(IFilePathToContentTypeProvider))]
    [Name("PabloYmlFilePathProvider")]
    [FileExtension(".yml")]
    internal sealed class PabloYmlFilePathProvider : PabloFilePathProviderBase
    {
    }

    internal static class PabloFileToContentTypeProvider
    {
        internal static bool IsPabloManifestFileName(string fileName)
        {
            if (fileName.Equals("pablo.yaml", StringComparison.OrdinalIgnoreCase) ||
                fileName.Equals("pablo.yml", StringComparison.OrdinalIgnoreCase))
            {
                return true;
            }

            if (!fileName.StartsWith("pablo", StringComparison.OrdinalIgnoreCase))
            {
                return false;
            }

            return fileName.EndsWith(".yaml", StringComparison.OrdinalIgnoreCase) ||
                   fileName.EndsWith(".yml", StringComparison.OrdinalIgnoreCase);
        }
    }
}
