#include "keyboard.h"

int key_event(unsigned char code,short up){
    int flag=up?2:0;
    keybd_event(code,0,flag,0);
    char *action=up?RELEASE:PRESS;
    printf("Cgo: %s key %d \n",action,code);
    return flag;
}
