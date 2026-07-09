using System;
using System.Collections.Generic;
using System.Text.RegularExpressions;

namespace Pablo.VisualStudio.Decorations
{
    internal enum StripeKind
    {
        Profile,
        Env,
    }

    internal sealed class MarkedLine
    {
        public int Line { get; set; }
        public string ColorKey { get; set; } = string.Empty;
        public int Indent { get; set; }
        public StripeKind Kind { get; set; }
    }

    internal static class ProfileStripeParser
    {
        private static readonly Regex MappingKeyRegex = new(@"^\s*([a-zA-Z0-9_-]+):\s*", RegexOptions.Compiled);

        public static IReadOnlyList<MarkedLine> Parse(string text)
        {
            var lines = text.Replace("\r\n", "\n").Split('\n');
            var marked = new List<MarkedLine>();
            var profilesIndent = -1;
            var currentProfile = string.Empty;
            var profileIndent = -1;
            var environmentsIndent = -1;
            var currentEnv = string.Empty;
            var envIndent = -1;

            for (var lineNum = 0; lineNum < lines.Length; lineNum++)
            {
                var line = lines[lineNum];
                var trimmed = line.Trim();
                var indent = GetIndent(line);
                var key = GetMappingKey(line);

                if (key == "profiles")
                {
                    profilesIndent = indent;
                    currentProfile = string.Empty;
                    profileIndent = -1;
                    environmentsIndent = -1;
                    currentEnv = string.Empty;
                    envIndent = -1;
                    continue;
                }

                if (profilesIndent < 0)
                {
                    continue;
                }

                if (!string.IsNullOrEmpty(trimmed) && indent <= profilesIndent)
                {
                    currentProfile = string.Empty;
                    profileIndent = -1;
                    environmentsIndent = -1;
                    currentEnv = string.Empty;
                    envIndent = -1;
                    continue;
                }

                if (!string.IsNullOrEmpty(key) && indent == profilesIndent + 2)
                {
                    currentProfile = key;
                    profileIndent = indent;
                    environmentsIndent = -1;
                    currentEnv = string.Empty;
                    envIndent = -1;
                    marked.Add(new MarkedLine
                    {
                        Line = lineNum,
                        ColorKey = $"profile:{currentProfile}",
                        Indent = profileIndent,
                        Kind = StripeKind.Profile,
                    });
                    continue;
                }

                if (string.IsNullOrEmpty(currentProfile) || profileIndent < 0)
                {
                    continue;
                }

                if (!string.IsNullOrEmpty(trimmed) && indent <= profileIndent)
                {
                    currentProfile = string.Empty;
                    profileIndent = -1;
                    environmentsIndent = -1;
                    currentEnv = string.Empty;
                    envIndent = -1;
                    continue;
                }

                if (key == "environments" && indent > profileIndent)
                {
                    environmentsIndent = indent;
                    currentEnv = string.Empty;
                    envIndent = -1;
                    marked.Add(new MarkedLine
                    {
                        Line = lineNum,
                        ColorKey = $"profile:{currentProfile}",
                        Indent = profileIndent,
                        Kind = StripeKind.Profile,
                    });
                    continue;
                }

                if (!string.IsNullOrEmpty(key) && environmentsIndent >= 0 && indent == environmentsIndent + 2)
                {
                    currentEnv = key;
                    envIndent = indent;
                    marked.Add(new MarkedLine
                    {
                        Line = lineNum,
                        ColorKey = $"profile:{currentProfile}",
                        Indent = profileIndent,
                        Kind = StripeKind.Profile,
                    });
                    marked.Add(new MarkedLine
                    {
                        Line = lineNum,
                        ColorKey = $"env:{currentProfile}/{currentEnv}",
                        Indent = envIndent,
                        Kind = StripeKind.Env,
                    });
                    continue;
                }

                if (!string.IsNullOrEmpty(currentEnv) && envIndent >= 0 && !string.IsNullOrEmpty(trimmed) && indent <= envIndent)
                {
                    currentEnv = string.Empty;
                    envIndent = -1;
                }

                if (environmentsIndent >= 0 && !string.IsNullOrEmpty(trimmed) && indent <= environmentsIndent && !string.IsNullOrEmpty(key))
                {
                    environmentsIndent = -1;
                    currentEnv = string.Empty;
                    envIndent = -1;
                }

                marked.Add(new MarkedLine
                {
                    Line = lineNum,
                    ColorKey = $"profile:{currentProfile}",
                    Indent = profileIndent,
                    Kind = StripeKind.Profile,
                });

                if (!string.IsNullOrEmpty(currentEnv) && envIndent >= 0 && (string.IsNullOrEmpty(trimmed) || indent > envIndent))
                {
                    marked.Add(new MarkedLine
                    {
                        Line = lineNum,
                        ColorKey = $"env:{currentProfile}/{currentEnv}",
                        Indent = envIndent,
                        Kind = StripeKind.Env,
                    });
                }
            }

            return marked;
        }

        private static int GetIndent(string line)
        {
            var match = Regex.Match(line, @"^(\s*)");
            return match.Success ? match.Groups[1].Value.Length : 0;
        }

        private static string? GetMappingKey(string line)
        {
            var match = MappingKeyRegex.Match(line);
            return match.Success ? match.Groups[1].Value : null;
        }
    }
}
