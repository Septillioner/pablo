using System.ComponentModel.Composition;
using Microsoft.VisualStudio.LanguageServer.Client;
using Microsoft.VisualStudio.Utilities;

namespace Pablo.VisualStudio.ContentType
{
    internal static class PabloContentTypeDefinitions
    {
        public const string ContentType = "pablo";
    }

    [Export]
    [Name(PabloContentTypeDefinitions.ContentType)]
    [BaseDefinition(CodeRemoteContentDefinition.CodeRemoteContentTypeName)]
    internal sealed class PabloContentTypeDefinition
    {
    }
}
