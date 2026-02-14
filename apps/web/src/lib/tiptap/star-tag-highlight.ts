import { Extension } from "@tiptap/core";
import { Decoration, DecorationSet } from "@tiptap/pm/view";
import { Plugin, PluginKey } from "@tiptap/pm/state";

/**
 * StarTagHighlight Extension
 * Highlights STAR tags ([상황], [과제], [행동], [결과]) in the editor
 */
export const StarTagHighlight = Extension.create({
  name: "starTagHighlight",

  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: new PluginKey("starTagHighlight"),
        props: {
          decorations(state) {
            const decorations: Decoration[] = [];
            const doc = state.doc;

            const starTags = [
              { pattern: /\[상황\]/g, class: "star-tag-situation" },
              { pattern: /\[과제\]/g, class: "star-tag-task" },
              { pattern: /\[행동\]/g, class: "star-tag-action" },
              { pattern: /\[결과\]/g, class: "star-tag-result" },
            ];

            doc.descendants((node, pos) => {
              if (node.isText && node.text) {
                const text = node.text;

                starTags.forEach(({ pattern, class: className }) => {
                  // Reset regex lastIndex
                  pattern.lastIndex = 0;
                  let match;

                  while ((match = pattern.exec(text)) !== null) {
                    const from = pos + match.index;
                    const to = from + match[0].length;

                    decorations.push(
                      Decoration.inline(from, to, {
                        class: className,
                      })
                    );
                  }
                });
              }
            });

            return DecorationSet.create(doc, decorations);
          },
        },
      }),
    ];
  },
});
