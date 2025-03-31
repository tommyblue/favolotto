//go:build pn7150

#ifndef NCF_LIB_H
#define NCF_LIB_H

typedef struct Tag {
  char uid[32];
  int uid_length;
  char text[1024];
  int text_length;
  int error;
} Tag;

void read_tag(void);

extern void exportTag(Tag *tag_context);
extern void onTagRemove(void);

//extern void tagUID(char *data, int lenght);
//extern void tagText(char *data, int lenght);

#endif // NCF_LIB_H
