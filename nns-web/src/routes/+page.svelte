<script lang="ts">
  import * as ucan from '@ucans/ucans'
	import { EdKeypair } from '@ucans/ucans';

  let privateKey: string = "lAgDDL4uR2JZ8dpGrPdINdLzOkkWOZSGT55f4wuk9kE+HwgFC9Oe879Ovv8b4RG1lEfwtOmamesPJ7RgEY/MSQ=="
  let name: string = "blaine@fission.codes"
  let delegationCID: string = ""
  let issuer: string = "did:web:bcook.ca"
  let audience: string = "did:mailto:blaine@fission.codes"
  let factType: string = "dkimProof"
  let factValue: string = ""
  let signedUcan: string = ""

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

  const myPlugin = new ucan.Plugins([], defaultRecord)
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

      pluginUcan.signWithKeypair(payload, keypair).then((signed) => {
        console.log(ucan, signed)
        signedUcan = ucan.encode(signed)
        console.log(signedUcan)
      })
    }
  }
</script>

<h1>NNS</h1>

<h2>Generate a UCAN</h2>

Name <input bind:value={name} placeholder="name" /><br/>
Issuer <input bind:value={issuer} /><br/>
Audience <input bind:value={audience} /><br/>

Fact type <input bind:value={factType} /><br/>
Fact value <input bind:value={factValue} /><br/>

Private Key <input bind:value={privateKey} type="password" /><br/>

Signed <textarea bind:value={signedUcan}></textarea>