<script lang="ts">

  import * as ucan from '@ucans/ucans'
	import { EdKeypair } from '@ucans/ucans';

  let privateKey: string = "lAgDDL4uR2JZ8dpGrPdINdLzOkkWOZSGT55f4wuk9kE+HwgFC9Oe879Ovv8b4RG1lEfwtOmamesPJ7RgEY/MSQ=="
  let didForPrivateKey = EdKeypair.fromSecretKey(privateKey).did()
  let name: string = "bcook.ca"
  let issuer: string = "did:dns:bcook.ca"
  let audience: string = "did:key:z6MkqUAn32xqHdUXsNGJ553zJHWw4E9SESc3naAYeomNZkic"
  let factType: string = "dnslink"
  let factValue: string = "QmdJBTuAM3JqiA2yhB32eHZiQ163XwMPYwdkL75fr7gJkM"
  let signedUcan: string = ""
  let ucanJSON: string = ""
  let stored: string = ""
  let storedBool: boolean = false

  const nullPlugin: ucan.DidMethodPlugin = {
    checkJwtAlg: (did: string, jwtAlg: string): boolean => {
      console.log('checking jwt alg', did, jwtAlg)
      return true
    },

    verifySignature: (did: string, data: Uint8Array, sig: Uint8Array): Promise<boolean> => {
      console.log('verifying signature', did, data, sig)
      throw new Error("Method not implemented.")
      return Promise.resolve(true)
    }
  }

  const defaultRecord: Record<string, ucan.DidMethodPlugin> = new Proxy({}, {
    get: (target, prop) => {
      console.log('getting', prop)
      return nullPlugin
    }
  })

  //const myPlugin = new ucan.Plugins([], defaultRecord)
  const myPlugin = new ucan.Plugins([], {"dns": nullPlugin})
  const pluginUcan = ucan.getPluginInjectedApi(myPlugin)

  $: {
    
    if (privateKey !== "" && name !== "") {
      
      const keypair = EdKeypair.fromSecretKey(privateKey)

      const [nameScheme, nameName] = name.split(':')

      let payload = ucan.buildPayload({
        lifetimeInSeconds: 60 * 60 * 24 * 365,
        audience,
        issuer,
        capabilities: [
          {
            with: { scheme: nameScheme, hierPart: nameName },
            can: { namespace: 'whocan', segments: ['be'] }
          }
        ],
        facts: [
          { type: factType, value: factValue }
        ]
      })

      ucanJSON = JSON.stringify(payload)

      pluginUcan.signWithKeypair(payload, keypair).then((signed) => {
        console.log(ucan, signed)
        signedUcan = ucan.encode(signed)
        console.log(signedUcan)
      })
    }
  }

  function handleClick () {
    fetch(`/${name}`, {
      method: 'POST',
      headers: {
        'Accept': 'application/json'
      },
      body: signedUcan
    }).then((res) => {
      if (res.status === 202) {
        stored = "Stored!"
        storedBool = true
      } else {
        res.text().then((text) => {
         stored = "Not stored! " +  text
        })
        storedBool = false
      }
      res.text().then((text) => {
        console.log(text)
      })
      return res
    })
  }
</script>

<style>
body {
    font-family: Arial, sans-serif;
    background-color: #f7f7f7;
    color: #333;
    line-height: 1.6;
    padding: 20px;
}

.container {
    max-width: 800px;
    margin: 0 auto;
    background-color: #fff;
    padding: 20px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

h1 {
    font-size: 2em;
    margin-bottom: 10px;
}

h2 {
    font-size: 1.5em;
    margin-bottom: 10px;
}

.input-group,
.textarea-group {
    display: flex;
    flex-direction: column;
    margin-bottom: 10px;
}

label {
    font-weight: bold;
    margin-bottom: 5px;
}

.input-field,
.textarea-field {
    padding: 8px;
    border: 1px solid #ccc;
    border-radius: 3px;
    font-size: 1em;
    outline: none;
}

.input-field:focus,
.textarea-field:focus {
    border-color: #0077cc;
    box-shadow: 0 0 4px rgba(0, 119, 204, 0.3);
}

.textarea-field {
    resize: vertical;
    min-height: 100px;
}

.submit-button {
    background-color: #0077cc;
    color: #fff;
    font-size: 1em;
    padding: 10px;
    border: none;
    border-radius: 3px;
    cursor: pointer;
    margin-top: 10px;
}

.submit-button:hover {
    background-color: #005fa3;
}

</style>

<body>
  <div class="container">
    <h1>NNS</h1>

    <h2>Generate a UCAN</h2>

    <div class="input-group">
        <label>Name</label>
        <input class="input-field" bind:value={name} placeholder="name" />
    </div>
    <div class="input-group">
        <label>Issuer</label>
        <input class="input-field" bind:value={issuer} />
    </div>
    <!-- <div class="input-group">
        <label>Audience</label>
        <input class="input-field" bind:value={audience} />
    </div> -->
    <div class="input-group">
        <label>Fact type</label>
        <input class="input-field" bind:value={factType} />
    </div>
    <div class="input-group">
        <label>Fact value</label>
        <input class="input-field" bind:value={factValue} />
    </div>
    <!-- <div class="input-group">
        <label>Private Key</label>
        <input class="input-field" bind:value={privateKey} type="password" />
    </div> -->
    <div class="input-group">
        <label>DID for Private Key</label>
        <input class="input-field" bind:value={didForPrivateKey} disabled />
    </div>

    <div class="textarea-group">
        <label>UCAN</label>
        <textarea class="textarea-field" bind:value={ucanJSON}></textarea>
    </div>

    <div class="textarea-group">
        <label>Signed</label>
        <textarea class="textarea-field" bind:value={signedUcan}></textarea>
    </div>

    <button class="submit-button" on:click={handleClick}>Store in NNS</button>
    <span>{stored} {#if storedBool}Fetch the UCAN here: <a href="/{name}">http://localhost:9970/{name}</a>{:else}{/if}</span>
  </div>

  <!-- <h1>NNS</h1>

  <h2>Generate a UCAN</h2>

  Name <input bind:value={name} placeholder="name" /><br/>
  Issuer <input bind:value={issuer} /><br/>
  Audience <input bind:value={audience} /><br/>

  Fact type <input bind:value={factType} /><br/>
  Fact value <input bind:value={factValue} /><br/>

  Private Key <input bind:value={privateKey} type="password" /><br/>
  DID for Private Key <input bind:value={didForPrivateKey} disabled /><br/>

  Signed <textarea bind:value={signedUcan}></textarea>

  Payload <textarea bind:value={ucanJSON}></textarea>

  <button on:click={handleClick}>Store in NNS</button> -->
</body>