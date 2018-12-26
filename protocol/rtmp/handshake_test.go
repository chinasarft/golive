package rtmp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"testing"
)

var (
	schema0c1 = "0000000009007c02f778551eceab8e1e362f07c5868a70b266d40220e5086118a22e3697d5fc2b6329c22ee7c7c9b0b5b345" +
		"dfa8398fea3bc879208a1068425aa1174cb57258cfebe24da42ec4ff9da19c7b406f18c95bfdd04ee48b6f833214a47b8bf6" +
		"8eb05254e25ae9a9a8b496c5496f6e9cd7c68cb6bcc84169e67f3e372ffe186da8513862c5afdeb5a1218a860baccdd869b2" +
		"53dbab8a0efb59a48c442099875e55b74fd07a1b69af303e099873cb77e2d9e2c4741eb1f506a7f0b2962af91f5051c890a3" +
		"20994f4c8b36a01fa9c0f15f4b7cf7b6c3817da60918611d3b02ff7cc94038937537d3382e416bd57577834df9caf7483521" +
		"50ce92d0dbb40e871fc364e18b7e3569e39876c1ae0cb79ff6d7ab32f9f011869c8da4191ef46d3c1bc6b72d83812f2243d1" +
		"752a412632d66a2075f5d302f45cdf2b98b802b79805dc4837d84eded1cc42cba7eec3137bf86dd0b2d7f624d521a3ff655b" +
		"cf0df4101341d230e94651628e45a7abf8183101ce981e5e86a6faec5c79aa7ec7181ea459ed966e0c932a6f9e4bf554ab7b" +
		"1c905cde054ed40d2fb89bbfc65752e58cfde4f2a04dc3ad1e5ef4fc64d57b071c14146a451f6a043156d2880f5957d78dfd" +
		"3ff2b880f624ac70f27094f37734eaf1924f66ac3d648ae232930544a4420b65eed4696f7f50fe05fc948516650bfb9b65f9" +
		"53d848e3456965d0d76b627a84e35e9ddabd8b43f0fadaee458cdf59ebd773922e0ae507d5e64092e90ec67ed8439e8f221c" +
		"753a9f9216f1cb85b5107bc7cc7c4fb2c3eed1afc709b192f50de69c1ffca5650efb523b5fab5997cac4b7ef5af2bf5973e6" +
		"293bae1fc1da499be5515077de4ddb3f7fbcd741711c62b7d803b8d054cf578182714d1c445bd692ba1b839a09e7266fc85d" +
		"aa8fad433dabbd206cde1ea571770c6bfa7956c266ee8435eb01caa6f583c0d83c33eefcf3f97db917bdb299f8a89fb8860a" +
		"b1cd644c1e82cce1323a9eb79d25f5dc570b70b40bb20061743353d0536ecdabd82f2277756228cb54fe6932c4c9bd130d2b" +
		"fc700bf5d85da6946e96f82e0afd8e9c20cfeabf9a6a189bb17bbef539d1d9e73ee966fcb5d6bdbf98444de865be1605e340" +
		"87b34cdf04e3aff55bd4137c522d2e3c7bf2634db326b22a474db7f8a0952ab3510de555396c18a378a8390b73bff7772fb2" +
		"760fa5786a3b0459c45a1d1b0dff4cbbb65da0ed34f32116c3c2cc2b95825111bcfbb246c1505aa00c970d06d76cc950ce36" +
		"8b220db20487cba02bb1fc54b36567b42913d6b885062cd7795601d31a4a8e6d56467fa5c52162676fd526122513c3e47f2e" +
		"6d9d6a19c4f8795431a94ced039221d27ac68a7a770e9305252271459b71d05505c4977b70e96851208bacdcdf7e682b3848" +
		"840880f1fec2a32fb1406862bc83cad89adf729f3ba6ffd3fe808e8403e7c6a0f6dcda3211898f6f16505b078cc78b80db60" +
		"0ded2f200f25ff5dcb31d32cbe2056b8bdc31091631abbd2ac28d61fe32480ec5035bbf35f173393f0d59a21d2300b2db4b5" +
		"9eaa63ffe0f93d65dc028516905f22a9c7dff387c4eb78c6300e7ebbb2a6663cfbe5a9977caf5764ba3f0aba57db4e6051e0" +
		"aec6e05b1bdaa649a97875edbc40bcc4214bb7fef72730d4368acbd3e10ab4c7642a1c53f93ca26db9c766555be26fdff0e2" +
		"a721425b95762b0b776fd3e3eb17313d53129c9f34003c99a5196a8c6aacb6d64b51dd98b27c5b59a70bd941a1ba0300fb74" +
		"4cfc499d997e6d93392d28a767d116dfbc2cb9eb40540d11eb304ad9322386fdc597d55d3d1c67bc54df35479afb19ea2cd2" +
		"04dd331efb49c1e172f6bb89fbf58207dcae6c12f8a7598bef1e872c0c35f22b4c29df10f9bb1add3aeb887ab5fb38c036b8" +
		"8346f761c42f295dfe0b437dfb4f30a71276ab7d409ce424963f9ebb4f1f47e52b18786f279b5a87226f66981dd6d0ad11f4" +
		"9af31055ff69710fda3457c4b7852db2e3936d0778bf879f1734ad3a24ffb1ae9584760b9cb0452a53ba0962ffe2f605af52" +
		"830bd211a5488894cc0b052255048711cd198510a9e943bf8b839198455fbd41073005d303990b88d9b63656d43cfec8ed83" +
		"748f4b0f0fc5120216794b22a054e5bc58abd8c41096070884393453ce509694afbeabe0"

	schema1c1 = "0000000009007c02189bb17bbef539d1d9e73ee966fcb5d6bdbf98444de865be1605e34087b34cdf04e3aff55bd4137c522d" +
		"2e3c7bf2634db326b22a474db7f8a0952ab3510de555396c18a378a8390b73bff7772fb2760fa5786a3b0459c45a1d1b0dff" +
		"4cbbb65da0ed34f32116c3c2cc2b95825111bcfbb246c1505aa00c970d06d76cc950ce368b220db20487cba02bb1fc54b365" +
		"67b42913d6b885062cd7795601d31a4a8e6d56467fa5c52162676fd526122513c3e47f2e6d9d6a19c4f8795431a94ced0392" +
		"21d27ac68a7a770e9305252271459b71d05505c4977b70e96851208bacdcdf7e682b3848840880f1fec2a32fb1406862bc83" +
		"cad89adf729f3ba6ffd3fe808e8403e7c6a0f6dcda3211898f6f16505b078cc78b80db600ded2f200f25ff5dcb31d32cbe20" +
		"56b8bdc31091631abbd2ac28d61fe32480ec5035bbf35f173393f0d59a21d2300b2db4b59eaa63ffe0f93d65dc028516905f" +
		"22a9c7dff387c4eb78c6300e7ebbb2a6663cfbe5a9977caf5764ba3f0aba57db4e6051e0aec6e05b1bdaa649a97875edbc40" +
		"bcc4214bb7fef72730d4368acbd3e10ab4c7642a1c53f93ca26db9c766555be26fdff0e2a721425b95762b0b776fd3e3eb17" +
		"313d53129c9f34003c99a5196a8c6aacb6d64b51dd98b27c5b59a70bd941a1ba0300fb744cfc499d997e6d93392d28a767d1" +
		"16dfbc2cb9eb40540d11eb304ad9322386fdc597d55d3d1c67bc54df35479afb19ea2cd204dd331efb49c1e172f6bb89fbf5" +
		"8207dcae6c12f8a7598bef1e872c0c35f22b4c29df10f9bb1add3aeb887ab5fb38c036b88346f761c42f295dfe0b437dfb4f" +
		"30a71276ab7d409ce424963f9ebb4f1f47e52b18786f279b5a87226f66981dd6d0ad11f49af31055ff69710fda3457c4b785" +
		"2db2e3936d0778bf879f1734ad3a24ffb1ae9584760b9cb0452a53ba0962ffe2f605af52830bd211a5488894cc0b05225504" +
		"8711cd198510a9e943bf8b839198455fbd41073005d303990b88d9b63656d43cfec8ed83748f4b0f0fc5120216794b22a054" +
		"e5bc58abd8c41096070884393453ce509694afbeabe0f778551eceab8e1e362f07c5868a70b266d40220e5086118a22e3697" +
		"d5fc2b6329c22ee7c7c9b0b5b345dfa8398fea3bc879208a1068425aa1174cb57258cfebe24da42ec4ff9da19c7b406f18c9" +
		"5bfdd04ee48b6f833214a47b8bf68eb05254e25ae9a9a8b496c5496f6e9cd7c68cb6bcc84169e67f3e372ffe186da8513862" +
		"c5afdeb5a1218a860baccdd869b253dbab8a0efb59a48c442099875e55b74fd07a1b69af303e099873cb77e2d9e2c4741eb1" +
		"f506a7f0b2962af91f5051c890a320994f4c8b36a01fa9c0f15f4b7cf7b6c3817da60918611d3b02ff7cc94038937537d338" +
		"2e416bd57577834df9caf748352150ce92d0dbb40e871fc364e18b7e3569e39876c1ae0cb79ff6d7ab32f9f011869c8da419" +
		"1ef46d3c1bc6b72d83812f2243d1752a412632d66a2075f5d302f45cdf2b98b802b79805dc4837d84eded1cc42cba7eec313" +
		"7bf86dd0b2d7f624d521a3ff655bcf0df4101341d230e94651628e45a7abf8183101ce981e5e86a6faec5c79aa7ec7181ea4" +
		"59ed966e0c932a6f9e4bf554ab7b1c905cde054ed40d2fb89bbfc65752e58cfde4f2a04dc3ad1e5ef4fc64d57b071c14146a" +
		"451f6a043156d2880f5957d78dfd3ff2b880f624ac70f27094f37734eaf1924f66ac3d648ae232930544a4420b65eed4696f" +
		"7f50fe05fc948516650bfb9b65f953d848e3456965d0d76b627a84e35e9ddabd8b43f0fadaee458cdf59ebd773922e0ae507" +
		"d5e64092e90ec67ed8439e8f221c753a9f9216f1cb85b5107bc7cc7c4fb2c3eed1afc709b192f50de69c1ffca5650efb523b" +
		"5fab5997cac4b7ef5af2bf5973e6293bae1fc1da499be5515077de4ddb3f7fbcd741711c62b7d803b8d054cf578182714d1c" +
		"445bd692ba1b839a09e7266fc85daa8fad433dabbd206cde1ea571770c6bfa7956c266ee8435eb01caa6f583c0d83c33eefc" +
		"f3f97db917bdb299f8a89fb8860ab1cd644c1e82cce1323a9eb79d25f5dc570b70b40bb20061743353d0536ecdabd82f2277" +
		"756228cb54fe6932c4c9bd130d2bfc700bf5d85da6946e96f82e0afd8e9c20cfeabf9a6a"

	s1 = "000000000d0e0a0d0dba3a2fc6e323f6bf51f19c722306a4723770c3b4605c8706e8629a3b4df631d3c75f3266af8d3b7f2d" +
		"163dff46447ceecb0ba289a46b80142bda88b96f7b947f1daf53f322f65b757bf0438a4172d79efe0846d11f9918eea98080" +
		"8c95e63d5c58a63ad30c421548ff2ffae46160320a7fae41202ba7261958ec1040c7ad26e328a011a556f39f0f4b5bd47da8" +
		"b219cc8d41ad0699ed443351c213b859bcdd6ff6d53ca260af51cf9c885a3c0f50205c6d78d3f21242bc188b2b864f0ca3ae" +
		"a224ad539ddec9338fcfac9b3c175b14501b2f563a3aa21fca5faa458121da9b0239da446a73cc951cfd4303deac33880a35" +
		"e2ca21c58c1a1941618d3aae29144fc961b0f8b9dc347402650543e495059250cd43d520194d12ae34d9a61de2d84dcf2600" +
		"207f91f91834cfa7f2caa4a241230e7b26d364e3dabb83b5b33f717ee5639db0f1f0853b2f96c9989832898ae933d94db4db" +
		"beea3e84e1ee71a389da9c4427792876e2538d6803c2c2d26b589c4862a34ab1bfad61275fb23e59505d870fd305c4076927" +
		"c022eeb340f8019cc68c47f8ab61ce02a279e3f53d1610a28ac2739704a8ac010d3b99928296b325999e624272d61eea2e0a" +
		"1ea50747451bdf8fb5efc653a571feb38b61a954ab0f400e5c17103ee4d7f9fc311792bc408a60090115fe9dd5117620aff8" +
		"66b115238ff509f026d20a19f4ac4396a07b161d4c1f2d1262c5e538a93e7fe91ea5b37ce8dfe7c0962adcb31b65be8ce38c" +
		"e4c728327386b14d4c86e1fed61b0fa10adac3a8ef3b7a1edb9c5b60552a0bf58840059d2edd5811697418bede98fda4f1e8" +
		"f34b1def91f284f1631595c712a2ef4f27070f54205ff99eddf2df4e04ce990e83cd973261d096a9006eb21aadd001f7bcd3" +
		"b3f746b99c5e4845977856c3211a8bf5cf2f1296a32bc7ce1150091f1682e8418a4f2c06a2ec10c699c6321969960b2512d0" +
		"f0ae59f80060b08d33c07b7001e43d0c33435b1294db9b95745771491a954bc00bff0b7c5199f642b6192a5f6a83c9bc9e31" +
		"5234c6813659a02a7d07e3711062bb412ee9d0d54c9a43470833e2de9c03c70c579b1fd6138ae627f17e98a50bbe245302b6" +
		"8d216eb339d1cd644e57c040992eaf406837b2292c9b5c18d2e7a2525e68d23f4bfff27954268eace60140450f1d495c76a5" +
		"b486b1091dce7394a76db1a98faf0b17ca9c88fb277cdb31daae25787b2a28345e78b3737c76286b39ff022ec1398183ab84" +
		"da7e175ce12cc0f590f1b08e72ad7b50cf2c5401e83de5b451673c4290d191bda3fbccb71450e6e6924e9f90efe99b6e4d02" +
		"8e15c82fc498ae7e7ad4a6aed63478a24dabb061834b30a930ba89784ed88c374124bf701b62b54ea8909f1030c9de611a78" +
		"74962e26e91ed9f1ebd29bd3396877845708fecd5ee4dc7cf318e9e108fc8c367690c4e49b7e1ab28c5a193148d1958fe67c" +
		"b3e2a1f50cc060d314761630c7c6f8e8c0c2c26891c024894d56a40992c731af052e825dc1a03d5fd1e0b136a5b3cb5877da" +
		"8dd99f65eb17a8f8d3fc06fc31916c00d424c2df64efa9ee8bff7179b39bae8d24bf2b00edbff630f866dc34bb22fe352be7" +
		"d602841a00b77b732cc3e2b10d6a683311916b3502cf0998dd6eba645f0875cfa0dc7f5179e9f271372fb2ff8794d1a480c1" +
		"d5c73770c4f79663e5733f23c647a7458cde43aa817a354bd0958bf2a0035513e2c4559008ad3af69d7dfd08d48c306502fb" +
		"8aaed180aa42f423148ed5666867579459bd5441077c483e74e586416c0843c22b04a2a2d2c57357f9fc246f7bbce74032d6" +
		"e5f4b260b461217b4e8a82446a0b938625eeafc71cc55ade4ba6b25e9801b4707a88d4b9721609e4413cc40689e7795566ee" +
		"15ea8cd2cc999cc02e263a14502668bcff9b3dee9a81458825bf25dba0e3839aa0d1159b3b014b83b801b8a2915ea3b5ac1d" +
		"69ed3fa852a7e8f79d249ff7e073178a02f71d4923dd3eb5bfbc48dde758063a9c45b6e4669cf84f557a6711204be5ea5a27" +
		"b5c1c90b750fe9a1ebb56b8d04fea3094af6339e7705611feaaaf4b62b43e0ffd299278bf7eeab4c3c4c89c6e2c69b389645" +
		"666ac1519a7eb2a1017d50fbc71fcd39cd4588141d83aaeed69921f9d89fdd6879a6730e"

	s2 = "7123241c0d112bdb9e09e80ae2a1fc2d0ebd69f6d7218c570a9c203e646f7c85d9a0f961bbc50bffb606a140b9a70cbe7674" +
		"2b531f033e177d0118b181bb1869de593d718c63c0a785fc1564157eb322aa377cc799927cbeb42159b563ae51f3ff934f3b" +
		"8a4b33c5f72a10e5f34f766936c793b942b995310d509aaefe6ff3fce121d877efe6c6f77a24f7e5070b908bb7cd73b9370c" +
		"4c07346dea8998a92ed6cb8357e22620ea2fa74133551e5337b37f7f5ff931088ca2e101ab031ca6f945e5d9c359ef3e3e43" +
		"fe9217662bb95cb5caf5bddb10dedd2801bbb5f76bb0cf3b406a4505612b711117f14a4189ab4ae5931c7471cb98d9d1ba02" +
		"c08ec9804011439321c02072419c02fded1c8f35057649cca01a35d97f60dbe1d1175829e84fde66d76a0c42cfe86a2f6409" +
		"5b689bf2988c55db1976d17de6def191b6f20465fa3ef3e232eac1601a2ed8656495d8c51f7ca5b34fd3d160e6e7a18f50a0" +
		"ddd872319e206b24f3e5f889abf293252a4b0c919fb0a86ce974404512c778ca24413617eb40f274587ad16c2e37ab281638" +
		"5762b53ffb2b844a329ca2534da61af8eb9d2c40fe726bb9f139dc3d1e2bd9c73234ac9151913f7fee745670cf5f0b9cb8e5" +
		"5ed6c14027c8604dd9263736b9175ba660892ac81f7fcd104fad70f6499263df6fbb03b425accd1929067c5ce2ac9daa0122" +
		"48bb7b6935050d579f2d09ad7dc7e36b6b5e688bcbd9753e36e7c83ffeeff1d8c50772fdab29a206eed437ba086c14d9113f" +
		"84098f5ba8b12d40780f89d56b4d020c269c11f59f046e8437e8c91a1ccc35e83fd7db4a8b86a85739152d2de0c92bba2231" +
		"7a97fc0d5fbb3698d6a69a797fc318480d74f1d6d78cca0137ffa386b19372b15f65a642ff5e0e22eb1351d5bf100bba8640" +
		"339020bcbbcde92555b06b9d574107ea1ce164b8b64691abd1f79c1541153eda39c797e247966947ce51ab58e8fab4da657a" +
		"c3f0fcae02cacd30e6e921bc975780a8dbe7708a47ae4fa08ea5f4cb4f6c2790021e872b8eea5c207eecd223ab48046171bd" +
		"fe3fd0224462a4350b696b3d07546910f918ae4c22718901ac4e20f91242b424c092afbc59dace3a121d2a80cf542c8adef6" +
		"a96a5bb4567b67ded36412660a276dc1d029cf888bec8a29a8339cfa44a49baf39bd52fd4eada5fae13a7f9c527814de3c5e" +
		"b31d148ba37843ac95c44cc32a0fc2026836f8730efbb5e73db846b7364a4786d071d6ee735d67aa9089dbe8b4d11e0224d7" +
		"52b0e83f75daf916ac7525cc2394c174dc768e94aafd0db8e1e0be4ebc7893e636d9c12d8836f2c74f5103b70cef1d22a084" +
		"9e3626a6653ee8c489d38582efeb7e01b336743f3303370396d07ad10096eb2eb4c27978036e95f80d627d448e58dbc28a2a" +
		"ef5af23d09d9418325349fae0767eda4e0ac745f8ea1c2b8a8c0508f77c2c6304a7c6f8729edb7aea223f43a8812ea6a1571" +
		"e5066985133f8c9479a6c111742b5ab8f5a5a32926fb4ecf3f8387b1c2c62a5482c534b6abb0d6af327383ab590f8d419b1c" +
		"6334aeace9509ddaafde3b8b91221f6544b6f6b98f5d5aa6c1c4643432dee6923c49a075ae327964beb3dbd29729ab24e402" +
		"840c6c40961c19f91cf9d4f6b7269fc2f6cef996757764d6f17ab268e294e42a62ba306de28e91e036bfb046b1a8d8ae8e76" +
		"e99e3216553fd8733cd1608bf928c3f32330b20fde39fc360fe1716c1208f974cc09db43159f216d7197439e0e48a32e7e79" +
		"7905de94439245a0c998ad1286000b4cc52a603206c84e4188685cd4a7449a09080447ad25c3bef94d9489b823175832df03" +
		"8c1e59f1a688433947c148e0e5cef59696b983dbdd09c6a0c7e970991e79a715f6b1ea19818b560d7c82f7cabcb55c16e638" +
		"ff0097af2a12fc1ccf9789e942d8bdf6024958bd5ce040065d75d42214afe44d294bb38136f1ee74707e31289a03f5852669" +
		"3d22b19d3ac04a2c36a52d5afc405c2e68ec0355e6df0cd8b57568ef5fb667fb7993278391cca674f796376e069d381a6c7a" +
		"e7848ea119186c6ca04159d7e0a3841c292ac247501b290df9bc4dbd4962c1abc0c1e717f464f2c49adc3ba52df853b99180" +
		"c7955acd375ee00dfb864f8dac57c0dd01cb06992cb64d098f33ddc943c288f94ff04fe4"

	testc0c1c2 = "030000000009007c02f778551eceab8e1e362f07c5868a70b266d40220e5086118a22e3697d5fc2b6329c22ee7c7c9b0b5b3" +
		"45dfa8398fea3bc879208a1068425aa1174cb57258cfebe24da42ec4ff9da19c7b406f18c95bfdd04ee48b6f833214a47b8b" +
		"f68eb05254e25ae9a9a8b496c5496f6e9cd7c68cb6bcc84169e67f3e372ffe186da8513862c5afdeb5a1218a860baccdd869" +
		"b253dbab8a0efb59a48c442099875e55b74fd07a1b69af303e099873cb77e2d9e2c4741eb1f506a7f0b2962af91f5051c890" +
		"a320994f4c8b36a01fa9c0f15f4b7cf7b6c3817da60918611d3b02ff7cc94038937537d3382e416bd57577834df9caf74835" +
		"2150ce92d0dbb40e871fc364e18b7e3569e39876c1ae0cb79ff6d7ab32f9f011869c8da4191ef46d3c1bc6b72d83812f2243" +
		"d1752a412632d66a2075f5d302f45cdf2b98b802b79805dc4837d84eded1cc42cba7eec3137bf86dd0b2d7f624d521a3ff65" +
		"5bcf0df4101341d230e94651628e45a7abf8183101ce981e5e86a6faec5c79aa7ec7181ea459ed966e0c932a6f9e4bf554ab" +
		"7b1c905cde054ed40d2fb89bbfc65752e58cfde4f2a04dc3ad1e5ef4fc64d57b071c14146a451f6a043156d2880f5957d78d" +
		"fd3ff2b880f624ac70f27094f37734eaf1924f66ac3d648ae232930544a4420b65eed4696f7f50fe05fc948516650bfb9b65" +
		"f953d848e3456965d0d76b627a84e35e9ddabd8b43f0fadaee458cdf59ebd773922e0ae507d5e64092e90ec67ed8439e8f22" +
		"1c753a9f9216f1cb85b5107bc7cc7c4fb2c3eed1afc709b192f50de69c1ffca5650efb523b5fab5997cac4b7ef5af2bf5973" +
		"e6293bae1fc1da499be5515077de4ddb3f7fbcd741711c62b7d803b8d054cf578182714d1c445bd692ba1b839a09e7266fc8" +
		"5daa8fad433dabbd206cde1ea571770c6bfa7956c266ee8435eb01caa6f583c0d83c33eefcf3f97db917bdb299f8a89fb886" +
		"0ab1cd644c1e82cce1323a9eb79d25f5dc570b70b40bb20061743353d0536ecdabd82f2277756228cb54fe6932c4c9bd130d" +
		"2bfc700bf5d85da6946e96f82e0afd8e9c20cfeabf9a6a189bb17bbef539d1d9e73ee966fcb5d6bdbf98444de865be1605e3" +
		"4087b34cdf04e3aff55bd4137c522d2e3c7bf2634db326b22a474db7f8a0952ab3510de555396c18a378a8390b73bff7772f" +
		"b2760fa5786a3b0459c45a1d1b0dff4cbbb65da0ed34f32116c3c2cc2b95825111bcfbb246c1505aa00c970d06d76cc950ce" +
		"368b220db20487cba02bb1fc54b36567b42913d6b885062cd7795601d31a4a8e6d56467fa5c52162676fd526122513c3e47f" +
		"2e6d9d6a19c4f8795431a94ced039221d27ac68a7a770e9305252271459b71d05505c4977b70e96851208bacdcdf7e682b38" +
		"48840880f1fec2a32fb1406862bc83cad89adf729f3ba6ffd3fe808e8403e7c6a0f6dcda3211898f6f16505b078cc78b80db" +
		"600ded2f200f25ff5dcb31d32cbe2056b8bdc31091631abbd2ac28d61fe32480ec5035bbf35f173393f0d59a21d2300b2db4" +
		"b59eaa63ffe0f93d65dc028516905f22a9c7dff387c4eb78c6300e7ebbb2a6663cfbe5a9977caf5764ba3f0aba57db4e6051" +
		"e0aec6e05b1bdaa649a97875edbc40bcc4214bb7fef72730d4368acbd3e10ab4c7642a1c53f93ca26db9c766555be26fdff0" +
		"e2a721425b95762b0b776fd3e3eb17313d53129c9f34003c99a5196a8c6aacb6d64b51dd98b27c5b59a70bd941a1ba0300fb" +
		"744cfc499d997e6d93392d28a767d116dfbc2cb9eb40540d11eb304ad9322386fdc597d55d3d1c67bc54df35479afb19ea2c" +
		"d204dd331efb49c1e172f6bb89fbf58207dcae6c12f8a7598bef1e872c0c35f22b4c29df10f9bb1add3aeb887ab5fb38c036" +
		"b88346f761c42f295dfe0b437dfb4f30a71276ab7d409ce424963f9ebb4f1f47e52b18786f279b5a87226f66981dd6d0ad11" +
		"f49af31055ff69710fda3457c4b7852db2e3936d0778bf879f1734ad3a24ffb1ae9584760b9cb0452a53ba0962ffe2f605af" +
		"52830bd211a5488894cc0b052255048711cd198510a9e943bf8b839198455fbd41073005d303990b88d9b63656d43cfec8ed" +
		"83748f4b0f0fc5120216794b22a054e5bc58abd8c41096070884393453ce509694afbeabe0000000000d0e0a0d0dba3a2fc6" +
		"e323f6bf51f19c722306a4723770c3b4605c8706e8629a3b4df631d3c75f3266af8d3b7f2d163dff46447ceecb0ba289a46b" +
		"80142bda88b96f7b947f1daf53f322f65b757bf0438a4172d79efe0846d11f9918eea980808c95e63d5c58a63ad30c421548" +
		"ff2ffae46160320a7fae41202ba7261958ec1040c7ad26e328a011a556f39f0f4b5bd47da8b219cc8d41ad0699ed443351c2" +
		"13b859bcdd6ff6d53ca260af51cf9c885a3c0f50205c6d78d3f21242bc188b2b864f0ca3aea224ad539ddec9338fcfac9b3c" +
		"175b14501b2f563a3aa21fca5faa458121da9b0239da446a73cc951cfd4303deac33880a35e2ca21c58c1a1941618d3aae29" +
		"144fc961b0f8b9dc347402650543e495059250cd43d520194d12ae34d9a61de2d84dcf2600207f91f91834cfa7f2caa4a241" +
		"230e7b26d364e3dabb83b5b33f717ee5639db0f1f0853b2f96c9989832898ae933d94db4dbbeea3e84e1ee71a389da9c4427" +
		"792876e2538d6803c2c2d26b589c4862a34ab1bfad61275fb23e59505d870fd305c4076927c022eeb340f8019cc68c47f8ab" +
		"61ce02a279e3f53d1610a28ac2739704a8ac010d3b99928296b325999e624272d61eea2e0a1ea50747451bdf8fb5efc653a5" +
		"71feb38b61a954ab0f400e5c17103ee4d7f9fc311792bc408a60090115fe9dd5117620aff866b115238ff509f026d20a19f4" +
		"ac4396a07b161d4c1f2d1262c5e538a93e7fe91ea5b37ce8dfe7c0962adcb31b65be8ce38ce4c728327386b14d4c86e1fed6" +
		"1b0fa10adac3a8ef3b7a1edb9c5b60552a0bf58840059d2edd5811697418bede98fda4f1e8f34b1def91f284f1631595c712" +
		"a2ef4f27070f54205ff99eddf2df4e04ce990e83cd973261d096a9006eb21aadd001f7bcd3b3f746b99c5e4845977856c321" +
		"1a8bf5cf2f1296a32bc7ce1150091f1682e8418a4f2c06a2ec10c699c6321969960b2512d0f0ae59f80060b08d33c07b7001" +
		"e43d0c33435b1294db9b95745771491a954bc00bff0b7c5199f642b6192a5f6a83c9bc9e315234c6813659a02a7d07e37110" +
		"62bb412ee9d0d54c9a43470833e2de9c03c70c579b1fd6138ae627f17e98a50bbe245302b68d216eb339d1cd644e57c04099" +
		"2eaf406837b2292c9b5c18d2e7a2525e68d23f4bfff27954268eace60140450f1d495c76a5b486b1091dce7394a76db1a98f" +
		"af0b17ca9c88fb277cdb31daae25787b2a28345e78b3737c76286b39ff022ec1398183ab84da7e175ce12cc0f590f1b08e72" +
		"ad7b50cf2c5401e83de5b451673c4290d191bda3fbccb71450e6e6924e9f90efe99b6e4d028e15c82fc498ae7e7ad4a6aed6" +
		"3478a24dabb061834b30a930ba89784ed88c374124bf701b62b54ea8909f1030c9de611a7874962e26e91ed9f1ebd29bd339" +
		"6877845708fecd5ee4dc7cf318e9e108fc8c367690c4e49b7e1ab28c5a193148d1958fe67cb3e2a1f50cc060d314761630c7" +
		"c6f8e8c0c2c26891c024894d56a40992c731af052e825dc1a03d5fd1e0b136a5b3cb5877da8dd99f65eb17a8f8d3fc06fc31" +
		"916c00d424c2df64efa9ee8bff7179b39bae8d24bf2b00edbff630f866dc34bb22fe352be7d602841a00b77b732cc3e2b10d" +
		"6a683311916b3502cf0998dd6eba645f0875cfa0dc7f5179e9f271372fb2ff8794d1a480c1d5c73770c4f79663e5733f23c6" +
		"47a7458cde43aa817a354bd0958bf2a0035513e2c4559008ad3af69d7dfd08d48c306502fb8aaed180aa42f423148ed56668" +
		"67579459bd5441077c483e74e586416c0843c22b04a2a2d2c57357f9fc246f7bbce74032d6e5f4b260b461217b4e8a82446a" +
		"0b938625eeafc71cc55ade4ba6b25e9801b4707a88d4b9721609e4413cc40689e7795566ee15ea8cd2cc999cc02e263a1450" +
		"2668bcff9b3dee9a81458825bf25dba0e3839aa0d1159b3b014b83b801b8a2915ea3b5ac1d69ed3fa852a7e8f79d249ff7e0" +
		"73178a02f71d4923dd3eb5bfbc48dde758063a9c45b6e4669cf84f557a6711204be5ea5a27b5c1c90b750fe9a1ebb56b8d04" +
		"fea3094af6339e7705611feaaaf4b62b43e0ffd299278bf7eeab4c3c4c89c6e2c69b389645666ac1519a7eb2a1017d50fbc7" +
		"1fcd39cd4588141d83aaeed69921f9d89fdd6879a6730e"

	tests0s1s2 = "03000000000d0e0a0d0dba3a2fc6e323f6bf51f19c722306a4723770c3b4605c8706e8629a3b4df631d3c75f3266af8d3b7f" +
		"2d163dff46447ceecb0ba289a46b80142bda88b96f7b947f1daf53f322f65b757bf0438a4172d79efe0846d11f9918eea980" +
		"808c95e63d5c58a63ad30c421548ff2ffae46160320a7fae41202ba7261958ec1040c7ad26e328a011a556f39f0f4b5bd47d" +
		"a8b219cc8d41ad0699ed443351c213b859bcdd6ff6d53ca260af51cf9c885a3c0f50205c6d78d3f21242bc188b2b864f0ca3" +
		"aea224ad539ddec9338fcfac9b3c175b14501b2f563a3aa21fca5faa458121da9b0239da446a73cc951cfd4303deac33880a" +
		"35e2ca21c58c1a1941618d3aae29144fc961b0f8b9dc347402650543e495059250cd43d520194d12ae34d9a61de2d84dcf26" +
		"00207f91f91834cfa7f2caa4a241230e7b26d364e3dabb83b5b33f717ee5639db0f1f0853b2f96c9989832898ae933d94db4" +
		"dbbeea3e84e1ee71a389da9c4427792876e2538d6803c2c2d26b589c4862a34ab1bfad61275fb23e59505d870fd305c40769" +
		"27c022eeb340f8019cc68c47f8ab61ce02a279e3f53d1610a28ac2739704a8ac010d3b99928296b325999e624272d61eea2e" +
		"0a1ea50747451bdf8fb5efc653a571feb38b61a954ab0f400e5c17103ee4d7f9fc311792bc408a60090115fe9dd5117620af" +
		"f866b115238ff509f026d20a19f4ac4396a07b161d4c1f2d1262c5e538a93e7fe91ea5b37ce8dfe7c0962adcb31b65be8ce3" +
		"8ce4c728327386b14d4c86e1fed61b0fa10adac3a8ef3b7a1edb9c5b60552a0bf58840059d2edd5811697418bede98fda4f1" +
		"e8f34b1def91f284f1631595c712a2ef4f27070f54205ff99eddf2df4e04ce990e83cd973261d096a9006eb21aadd001f7bc" +
		"d3b3f746b99c5e4845977856c3211a8bf5cf2f1296a32bc7ce1150091f1682e8418a4f2c06a2ec10c699c6321969960b2512" +
		"d0f0ae59f80060b08d33c07b7001e43d0c33435b1294db9b95745771491a954bc00bff0b7c5199f642b6192a5f6a83c9bc9e" +
		"315234c6813659a02a7d07e3711062bb412ee9d0d54c9a43470833e2de9c03c70c579b1fd6138ae627f17e98a50bbe245302" +
		"b68d216eb339d1cd644e57c040992eaf406837b2292c9b5c18d2e7a2525e68d23f4bfff27954268eace60140450f1d495c76" +
		"a5b486b1091dce7394a76db1a98faf0b17ca9c88fb277cdb31daae25787b2a28345e78b3737c76286b39ff022ec1398183ab" +
		"84da7e175ce12cc0f590f1b08e72ad7b50cf2c5401e83de5b451673c4290d191bda3fbccb71450e6e6924e9f90efe99b6e4d" +
		"028e15c82fc498ae7e7ad4a6aed63478a24dabb061834b30a930ba89784ed88c374124bf701b62b54ea8909f1030c9de611a" +
		"7874962e26e91ed9f1ebd29bd3396877845708fecd5ee4dc7cf318e9e108fc8c367690c4e49b7e1ab28c5a193148d1958fe6" +
		"7cb3e2a1f50cc060d314761630c7c6f8e8c0c2c26891c024894d56a40992c731af052e825dc1a03d5fd1e0b136a5b3cb5877" +
		"da8dd99f65eb17a8f8d3fc06fc31916c00d424c2df64efa9ee8bff7179b39bae8d24bf2b00edbff630f866dc34bb22fe352b" +
		"e7d602841a00b77b732cc3e2b10d6a683311916b3502cf0998dd6eba645f0875cfa0dc7f5179e9f271372fb2ff8794d1a480" +
		"c1d5c73770c4f79663e5733f23c647a7458cde43aa817a354bd0958bf2a0035513e2c4559008ad3af69d7dfd08d48c306502" +
		"fb8aaed180aa42f423148ed5666867579459bd5441077c483e74e586416c0843c22b04a2a2d2c57357f9fc246f7bbce74032" +
		"d6e5f4b260b461217b4e8a82446a0b938625eeafc71cc55ade4ba6b25e9801b4707a88d4b9721609e4413cc40689e7795566" +
		"ee15ea8cd2cc999cc02e263a14502668bcff9b3dee9a81458825bf25dba0e3839aa0d1159b3b014b83b801b8a2915ea3b5ac" +
		"1d69ed3fa852a7e8f79d249ff7e073178a02f71d4923dd3eb5bfbc48dde758063a9c45b6e4669cf84f557a6711204be5ea5a" +
		"27b5c1c90b750fe9a1ebb56b8d04fea3094af6339e7705611feaaaf4b62b43e0ffd299278bf7eeab4c3c4c89c6e2c69b3896" +
		"45666ac1519a7eb2a1017d50fbc71fcd39cd4588141d83aaeed69921f9d89fdd6879a6730e7123241c0d112bdb9e09e80ae2" +
		"a1fc2d0ebd69f6d7218c570a9c203e646f7c85d9a0f961bbc50bffb606a140b9a70cbe76742b531f033e177d0118b181bb18" +
		"69de593d718c63c0a785fc1564157eb322aa377cc799927cbeb42159b563ae51f3ff934f3b8a4b33c5f72a10e5f34f766936" +
		"c793b942b995310d509aaefe6ff3fce121d877efe6c6f77a24f7e5070b908bb7cd73b9370c4c07346dea8998a92ed6cb8357" +
		"e22620ea2fa74133551e5337b37f7f5ff931088ca2e101ab031ca6f945e5d9c359ef3e3e43fe9217662bb95cb5caf5bddb10" +
		"dedd2801bbb5f76bb0cf3b406a4505612b711117f14a4189ab4ae5931c7471cb98d9d1ba02c08ec9804011439321c0207241" +
		"9c02fded1c8f35057649cca01a35d97f60dbe1d1175829e84fde66d76a0c42cfe86a2f64095b689bf2988c55db1976d17de6" +
		"def191b6f20465fa3ef3e232eac1601a2ed8656495d8c51f7ca5b34fd3d160e6e7a18f50a0ddd872319e206b24f3e5f889ab" +
		"f293252a4b0c919fb0a86ce974404512c778ca24413617eb40f274587ad16c2e37ab2816385762b53ffb2b844a329ca2534d" +
		"a61af8eb9d2c40fe726bb9f139dc3d1e2bd9c73234ac9151913f7fee745670cf5f0b9cb8e55ed6c14027c8604dd9263736b9" +
		"175ba660892ac81f7fcd104fad70f6499263df6fbb03b425accd1929067c5ce2ac9daa012248bb7b6935050d579f2d09ad7d" +
		"c7e36b6b5e688bcbd9753e36e7c83ffeeff1d8c50772fdab29a206eed437ba086c14d9113f84098f5ba8b12d40780f89d56b" +
		"4d020c269c11f59f046e8437e8c91a1ccc35e83fd7db4a8b86a85739152d2de0c92bba22317a97fc0d5fbb3698d6a69a797f" +
		"c318480d74f1d6d78cca0137ffa386b19372b15f65a642ff5e0e22eb1351d5bf100bba8640339020bcbbcde92555b06b9d57" +
		"4107ea1ce164b8b64691abd1f79c1541153eda39c797e247966947ce51ab58e8fab4da657ac3f0fcae02cacd30e6e921bc97" +
		"5780a8dbe7708a47ae4fa08ea5f4cb4f6c2790021e872b8eea5c207eecd223ab48046171bdfe3fd0224462a4350b696b3d07" +
		"546910f918ae4c22718901ac4e20f91242b424c092afbc59dace3a121d2a80cf542c8adef6a96a5bb4567b67ded36412660a" +
		"276dc1d029cf888bec8a29a8339cfa44a49baf39bd52fd4eada5fae13a7f9c527814de3c5eb31d148ba37843ac95c44cc32a" +
		"0fc2026836f8730efbb5e73db846b7364a4786d071d6ee735d67aa9089dbe8b4d11e0224d752b0e83f75daf916ac7525cc23" +
		"94c174dc768e94aafd0db8e1e0be4ebc7893e636d9c12d8836f2c74f5103b70cef1d22a0849e3626a6653ee8c489d38582ef" +
		"eb7e01b336743f3303370396d07ad10096eb2eb4c27978036e95f80d627d448e58dbc28a2aef5af23d09d9418325349fae07" +
		"67eda4e0ac745f8ea1c2b8a8c0508f77c2c6304a7c6f8729edb7aea223f43a8812ea6a1571e5066985133f8c9479a6c11174" +
		"2b5ab8f5a5a32926fb4ecf3f8387b1c2c62a5482c534b6abb0d6af327383ab590f8d419b1c6334aeace9509ddaafde3b8b91" +
		"221f6544b6f6b98f5d5aa6c1c4643432dee6923c49a075ae327964beb3dbd29729ab24e402840c6c40961c19f91cf9d4f6b7" +
		"269fc2f6cef996757764d6f17ab268e294e42a62ba306de28e91e036bfb046b1a8d8ae8e76e99e3216553fd8733cd1608bf9" +
		"28c3f32330b20fde39fc360fe1716c1208f974cc09db43159f216d7197439e0e48a32e7e797905de94439245a0c998ad1286" +
		"000b4cc52a603206c84e4188685cd4a7449a09080447ad25c3bef94d9489b823175832df038c1e59f1a688433947c148e0e5" +
		"cef59696b983dbdd09c6a0c7e970991e79a715f6b1ea19818b560d7c82f7cabcb55c16e638ff0097af2a12fc1ccf9789e942" +
		"d8bdf6024958bd5ce040065d75d42214afe44d294bb38136f1ee74707e31289a03f58526693d22b19d3ac04a2c36a52d5afc" +
		"405c2e68ec0355e6df0cd8b57568ef5fb667fb7993278391cca674f796376e069d381a6c7ae7848ea119186c6ca04159d7e0" +
		"a3841c292ac247501b290df9bc4dbd4962c1abc0c1e717f464f2c49adc3ba52df853b99180c7955acd375ee00dfb864f8dac" +
		"57c0dd01cb06992cb64d098f33ddc943c288f94ff04fe4"
)

type testHandshakeer struct {
	r         io.Reader
	writeChan chan int
	writeBuf  []byte
}

func newTestHandshakeer(c0c1c2 []byte) *testHandshakeer {
	return &testHandshakeer{
		r:         bytes.NewReader(c0c1c2),
		writeBuf:  make([]byte, 0, 1024*5),
		writeChan: make(chan int),
	}
}

func (hs *testHandshakeer) Read(p []byte) (n int, err error) {
	return hs.r.Read(p)
}

func (hs *testHandshakeer) Write(p []byte) (n int, err error) {
	copy(hs.writeBuf[len(hs.writeBuf):], p)
	return len(p), nil
}

func checkC1(t *testing.T, strc1 string) error {
	c1 := make([]byte, len(strc1)/2)
	len, err := hex.Decode(c1, []byte(strc1))
	if err != nil {
		return fmt.Errorf("hex decode c1 fail:%s", err)
	}
	if len != 1536 {
		return fmt.Errorf("wrong c1 data:%d", len)
	}

	ok, _ := complexHandshakeC1CheckAndDigest(c1)
	if !ok {
		return fmt.Errorf("check C1 digest fail:")
	}
	return nil
}

func TestCheckC1(t *testing.T) {
	err := checkC1(t, schema0c1)
	if err != nil {
		t.Error(err)
	} else {
		log.Println("schema0 success")
	}

	// err = checkC1(t, schema1c1)
	// if err != nil {
	// 	t.Error(err)
	// }
}

func TestHandshakeServer(t *testing.T) {
	c0c1c2 := make([]byte, len(testc0c1c2)/2)
	_, err := hex.Decode(c0c1c2, []byte(testc0c1c2))
	if err != nil {
		t.Errorf("hex decode c1 fail:%s", err)
	}

	hs := newTestHandshakeer(c0c1c2)

	err = handshakeServer(hs)
	if err != nil {
		t.Errorf("check C1 digest fail:")
	} else {
		log.Println("handshake ok")
	}
	return
}

type testConn struct {
	achan chan byte
	bchan chan byte
}

func newTestConn(achan, bchan chan byte) *testConn {
	return &testConn{
		achan: achan,
		bchan: bchan,
	}
}

func (c *testConn) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i] = <-c.achan
	}
	log.Println("read", len(p), "byte")
	return len(p), nil
}

func (c *testConn) Write(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		c.bchan <- p[i]
	}
	log.Println("write", len(p), "byte")
	return len(p), nil
}

func wrapHandshakeClient(c1 io.ReadWriter, errchan chan error) {
	err := HandshakeClient(c1)
	if err == nil {
		log.Println("handshake client ok")
		errchan <- nil
	} else {
		log.Println("handshake client fail:", err.Error())
		errchan <- err
	}
}
func TestHandshakeClient(t *testing.T) {
	achan := make(chan byte, 1024*1024)
	bchan := make(chan byte, 1024*1024)
	c1 := newTestConn(achan, bchan)
	c2 := newTestConn(bchan, achan)

	errchan := make(chan error)

	go wrapHandshakeClient(c1, errchan)

	err := handshakeServer(c2)
	if err == nil {
		log.Println("handshake server ok")
	} else {
		t.Error("handshake server fail:", err.Error())
	}

	err = <-errchan

	if err != nil {
		t.Error("handshake client ok")
	}
}
