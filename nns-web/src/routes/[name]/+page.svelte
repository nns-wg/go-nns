<script lang="ts">
  import type { UcanParts } from '@ucans/ucans'
  import * as ucans from '@ucans/ucans'

  const formatJson = (maybeJson: Record<string, unknown> | undefined, lineWidth: number = 0): string => {

    const jsonStr = maybeJson !== undefined ?  JSON.stringify(maybeJson, null, 2) : ''

    if (lineWidth === 0) {
      return jsonStr;
    }

    const lines = jsonStr.split('\n');
    const wrappedLines: string[] = [];

    lines.forEach(line => {
      let remainingLine = line;

      while (remainingLine.length > lineWidth) {
        const lastSpaceIndex = remainingLine.slice(0, lineWidth).lastIndexOf(' ');
        const wrapAt = lastSpaceIndex === -1 ? lineWidth : lastSpaceIndex;
        wrappedLines.push(remainingLine.slice(0, wrapAt));
        remainingLine = remainingLine.slice(wrapAt + 1);
      }

      wrappedLines.push(remainingLine);
    });

    return wrappedLines.join('\n');
  }

  import { CodeSnippet } from 'carbon-components-svelte'

  export let data
</script>

{#if data.error}

  <p>{data.error.message}</p>

{:else if data.ucan}

<CodeSnippet hideCopyButton>
<!-- {formatJson(decodedUcan?.header)} -->
  {formatJson(data.ucan.payload, 80)}
</CodeSnippet>

{/if}