#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <ctype.h>
#include <string.h>
#include <pthread.h>

#include "linux_nfc_api.h"
#include "_cgo_export.h"

pthread_cond_t condition = PTHREAD_COND_INITIALIZER;
pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;

nfcTagCallback_t g_TagCB;
nfc_tag_info_t g_tagInfos;

void onTagArrival(nfc_tag_info_t *pTagInfo){
  printf("Tag detected\n");
  g_tagInfos = *pTagInfo;
  pthread_cond_signal(&condition);
}

void read_tag(void) {
  ndef_info_t NDEFinfo;
  unsigned char* NDEFContent = NULL;
  nfc_friendly_type_t lNDEFType = NDEF_FRIENDLY_TYPE_OTHER;

  Tag tag_context;
  memset((void *)&tag_context, 0x0, sizeof(Tag));

  g_TagCB.onTagArrival = onTagArrival;
  g_TagCB.onTagDeparture = onTagRemove;

  nfcManager_doInitialize();
  nfcManager_registerTagCallback(&g_TagCB);
  nfcManager_enableDiscovery(DEFAULT_NFA_TECH_MASK, 0x01, 0, 0);

  printf("Waiting for tag...\n");
  do{
    int res = 0;
    memset((void *)&tag_context, 0x0, sizeof(Tag));

    pthread_cond_wait(&condition, &mutex);
    memcpy((void *)&tag_context.uid, g_tagInfos.uid, g_tagInfos.uid_length);

    res = nfcTag_isNdef(g_tagInfos.handle, &NDEFinfo);
    tag_context.error = 1; //Not a NDEF tag

    if(res == 1) {
      NDEFContent = malloc(NDEFinfo.current_ndef_length * sizeof(unsigned char));
      res = nfcTag_readNdef(g_tagInfos.handle, NDEFContent, NDEFinfo.current_ndef_length, &lNDEFType);

      tag_context.error = 2; //Not a NDEF Text record
      if(lNDEFType == NDEF_FRIENDLY_TYPE_TEXT) {
        res = ndef_readText(NDEFContent, res, tag_context.text, res);
        tag_context.error = 3; //Read NDEF Text Error
        if(res >= 0) {
          tag_context.error = 0;
          tag_context.text_length = res;
        }
      }
    }
    exportTag(&tag_context);
  } while(1);

  nfcManager_doDeinitialize();
}
