import bip39 from "bip39";
import { hdkey } from "ethereumjs-wallet";
import { bufferToHex, pubToAddress } from "ethereumjs-util";
import { Wallet } from "ethers";

//mnemonic → seed → HD wallet derivation → private key → public key → address
export const CreateWallet = async () => {
  // 1. create mnemonic
  const mnemonic = bip39.generateMnemonic();

  // 2. generate seed from mnemonic
  const seed = await bip39.mnemonicToSeed(mnemonic);

  // 3. derive HD wallet (Ethereum path)
  const hd = hdkey.fromMasterSeed(seed);
  const node = hd.derivePath("m/44'/60'/0'/0/0");
  const wallet = node.getWallet();

  // 4. get private key
  const privateKey = wallet.getPrivateKey().toString("hex");

  // 5. get public key
  const publicKey = bufferToHex(wallet.getPublicKey());

  // 6. get address
  const addressBuffer = pubToAddress(wallet.getPublicKey(), true);
  const address = `0x${addressBuffer.toString("hex")}`;

  return {
    mnemonic,
    privateKey,
    publicKey,
    address,
  };
};

export const WalletFromMnemonic = (mnemonic: string) => {
  const pharse = mnemonic.trim();
  const seed = bip39.mnemonicToSeedSync(pharse);
  const hd = hdkey.fromMasterSeed(seed);
  const node = hd.derivePath("m/44'/60'/0'/0/0");
  const ethWallet = node.getWallet();

  const privateKey = ethWallet.getPrivateKey().toString("hex");
  return new Wallet(`0x${privateKey}`);
};
