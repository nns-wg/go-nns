import * as ucans from '@ucans/ucans'


interface ParsedEmail {
  headers: { [key: string]: string };
  body: string;
}

export async function load({params}) {
  const response = await fetch(`/${params.name}`, {headers: [["accept", "application/json"]]})
  try {
    const text = await response.text()
    console.log(text);
    if (response.status === 200) {
      const ucan = ucans.parse(text)
      if (!ucan) {
        return {
          error: {
            code: 500,
            message: 'Invalid UCAN'
          }
        }
      }

      return {
        ucan: prettifyUcan(ucan)
      }

    } else if (response.status === 404) {
      return {
        error: {
          code: 404,
          message: `${params.name} not found`
        }
      }

    }
  } catch(e) {
    console.error(e)
    return {
      error: {
        code: 500,
        message: e
      }
    }
  }
}

function prettifyUcan(ucan: ucans.UcanParts): ucans.UcanParts {
  const idx = ucan.payload.fct?.findIndex(fct => fct['dkimProof'])
  if (ucan.payload.fct && idx !== undefined && idx !== -1) {
    const dkimProof = ucan.payload.fct[idx]
    if (dkimProof !== undefined) {
      const parsed = parseRfc2822Email(dkimProof['dkimProof'])
      ucan.payload.fct[idx] = parsed
    }
  }

  return ucan
}

function parseRfc2822Email(rawEmail: string): ParsedEmail {
  console.log(rawEmail)
  const lines = rawEmail.split(/\r?\n/);
  const headers: { [key: string]: string } = {};
  let body = '';
  let isHeader = true;

  lines.forEach((line: string) => {
    if (isHeader && line === '') {
      isHeader = false;
      return;
    }

    if (isHeader) {
      const headerMatch = line.match(/^([\w-]+):\s*(.*)$/);

      if (headerMatch) {
        const key = headerMatch[1];
        const value = headerMatch[2];
        headers[key] = value;
      } else {
        console.warn(`Invalid header: ${line}`);
      }
    } else {
      body += line + '\n';
    }
  });

  return { headers, body };
}