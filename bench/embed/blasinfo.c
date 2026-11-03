// Prints which kernel set a DYNAMIC_ARCH OpenBLAS picked on this host.
#include <stdio.h>
char *openblas_get_config(void);
char *openblas_get_corename(void);
int main(void) {
    printf("openblas_config=%s\nopenblas_corename=%s\n", openblas_get_config(), openblas_get_corename());
    return 0;
}
