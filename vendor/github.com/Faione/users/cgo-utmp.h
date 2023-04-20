#ifndef CGO_UTMP_H_
#define CGO_UTMP_H_

#include <utmp.h>
#include <stdlib.h>
#include <stdio.h>

// Define a function to open the utmp file
FILE* cgo_utmp_open() {
  return fopen("/var/run/utmp", "r");
}

// Define a function to read the next entry from the utmp file
int cgo_utmp_next(FILE* fp, struct utmp* ut) {
  return fread(ut, sizeof(struct utmp), 1, fp);
}

// Define a function to close the utmp file
void cgo_utmp_close(FILE* fp) {
  fclose(fp);
}

#endif // CGO_UTMP_H_