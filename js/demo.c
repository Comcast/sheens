#include <string.h>
#include <stdio.h>

#include "duktape.h"
#include "sheens_js.h"

typedef struct {
  duk_context *dctx;
} Ctx;

static Ctx *ctx = NULL;

void *make_ctx() {
  void* ret = malloc(sizeof(Ctx));

  memset(ret, 0, sizeof(Ctx));

  return ret;
}

char* readFile(const char *filename) {
  fprintf(stderr, "reading '%s'\n", filename);
  
  char * buffer = 0;
  long length;
  FILE * f = fopen(filename, "rb");
  
  if (f) {
    fseek(f, 0, SEEK_END);
    length = ftell(f);
    fseek(f, 0, SEEK_SET);
    buffer = malloc(length);
    if (buffer) {
      long int n = fread(buffer, 1, length, f);
      if (n != length) {
	fprintf(stderr, "warning: readFile %s read %ld != %ld", filename, n, length);
      }
    }
    fclose(f);
  } else {
    fprintf(stderr, "couldn't read '%s'\n", filename);
    exit(1);
  }

  buffer[length-1] = 0;

  fprintf(stderr, "read %ld bytes from '%s'\n", length, filename);

  return buffer;
}

static duk_ret_t native_print(duk_context *ctx) {
	duk_push_string(ctx, " ");
	duk_insert(ctx, 0);
	duk_join(ctx, duk_get_top(ctx) - 1);
	printf("%s\n", duk_to_string(ctx, -1));
	return 0;
}

static duk_ret_t eval_raw(duk_context *ctx, void *udata) {
	(void) udata;
	duk_eval(ctx);
	return 1;
}

static duk_ret_t tostring_raw(duk_context *ctx, void *udata) {
	(void) udata;
	duk_to_string(ctx, -1);
	return 1;
}

static void usage_exit(void) {
	fprintf(stderr, "Usage: eval <expression> [<expression>] ...\n");
	fflush(stderr);
	exit(1);
}

char * strdup(const char *s) {
  size_t n;
  char *acc;

  if (s == NULL) return NULL;

  n = strlen(s) + 1;
  acc = (char*) malloc(n);
  if (acc == (char*) 0) {
    return (char*) 0;
  }
  return (char*) memcpy(acc, s, n);
}

static duk_ret_t sandbox(duk_context *ctx) {
  const char *src = duk_to_string(ctx, 0);

  duk_context *box = duk_create_heap_default();
  duk_push_string(box, src);

  duk_ret_t rc = duk_peval(box);
  /* If we ran into an error, it's on the stack. */
  const char *result = duk_safe_to_string(box, -1);
  result = strdup(result);
  if (rc != DUK_EXEC_SUCCESS) {
    fprintf(stderr, "warning: sandbox returned non-zero rc=%d result=%s code:\n%s\n", rc, result, src);
  }

  duk_destroy_heap(box);
  duk_push_string(ctx, result);
  free((char*) result); /* result was interned! */

  return 1; /* If non-zero, caller will see 'undefined'. */
}

static duk_ret_t readfile(duk_context *ctx) {
  const char *filename = duk_to_string(ctx, 0);
  char *buf = readFile(filename);
  duk_push_string(ctx, buf);

  return 1; /* If non-zero, caller will see 'undefined'. */
}

int main(int argc, char *argv[]) {

	const char *res;

	if (argc < 2) {
		usage_exit();
	}

	ctx = make_ctx();
	ctx->dctx = duk_create_heap_default();
	
	duk_push_c_function(ctx->dctx, native_print, DUK_VARARGS);
	duk_put_global_string(ctx->dctx, "print");

	duk_push_c_function(ctx->dctx, sandbox, 1);
	duk_put_global_string(ctx->dctx, "sandbox");

	duk_push_c_function(ctx->dctx, readfile, 1);
	duk_put_global_string(ctx->dctx, "readfile");

	{
	  char *src = sheens_js();
	  {
	    int rc = duk_peval_string(ctx->dctx, src);
	    if (rc != 0) {
	      exit(rc);
	    }
	  }
	}

	{
	  int i;
	  for (i = 1; i < argc; i++) {
	    
	    char *buf = readFile(argv[i]);
	    duk_push_string(ctx->dctx, buf);
	    duk_safe_call(ctx->dctx, eval_raw, NULL, 1 /*nargs*/, 1 /*nrets*/);
	    duk_safe_call(ctx->dctx, tostring_raw, NULL, 1 /*nargs*/, 1 /*nrets*/);
	    res = duk_get_string(ctx->dctx, -1);
	    printf("%s\n", res ? res : "null");
	    duk_pop(ctx->dctx);
	  }
	}

	duk_destroy_heap(ctx->dctx);

	return 0;
}
