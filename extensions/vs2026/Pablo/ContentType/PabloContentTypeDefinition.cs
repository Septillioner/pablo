using System.ComponentModel.Composition;
using Microsoft.VisualStudio.LanguageServer.Client;
using Microsoft.VisualStudio.Utilities;

namespace Pablo.VisualStudio.ContentType
{
    internal static class PabloContentTypeDefinitions
    {
        public const string ContentType = "pablo";

        // ContentTypeDefinition is sealed — export a field of that type, not a custom class.
        // Exporting a non-ContentTypeDefinition type means "pablo" never registers, so
        // IFilePathToContentTypeProvider / ILanguageClient never activate for manifests.
        [Export]
        [Name(ContentType)]
        [BaseDefinition(CodeRemoteContentDefinition.CodeRemoteContentTypeName)]
        internal static ContentTypeDefinition PabloContentTypeDefinition = null!;
    }
}
