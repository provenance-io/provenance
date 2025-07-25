package io.provenance.client.protobuf.extensions

import cosmos.auth.v1beta1.QueryOuterClass
import io.provenance.marker.v1.MarkerAccount
import cosmos.auth.v1beta1.QueryGrpc.QueryBlockingStub as BlockingAuth
import cosmos.auth.v1beta1.QueryGrpcKt.QueryCoroutineStub as CoroutineAuth
import cosmos.vesting.v1beta1.Vesting

/**
 * Given an address, get the base account associated with it.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L18).
 *
 * @param bech32Address The bech32 address to fetch.
 * @return [cosmos.auth.v1beta1.Auth.BaseAccount] or throw [IllegalArgumentException] if the account type is not supported.
 */
fun BlockingAuth.getBaseAccount(bech32Address: String): cosmos.auth.v1beta1.Auth.BaseAccount =
    account(QueryOuterClass.QueryAccountRequest.newBuilder().setAddress(bech32Address).build()).account.run {
        when {
            this.`is`(cosmos.auth.v1beta1.Auth.BaseAccount::class.java) -> unpack(cosmos.auth.v1beta1.Auth.BaseAccount::class.java)
            this.`is`(Vesting.BaseVestingAccount::class.java) -> unpack(Vesting.BaseVestingAccount::class.java).baseAccount
            this.`is`(Vesting.ContinuousVestingAccount::class.java) -> unpack(Vesting.ContinuousVestingAccount::class.java).baseVestingAccount.baseAccount
            this.`is`(Vesting.DelayedVestingAccount::class.java) -> unpack(Vesting.DelayedVestingAccount::class.java).baseVestingAccount.baseAccount
            this.`is`(Vesting.PeriodicVestingAccount::class.java) -> unpack(Vesting.PeriodicVestingAccount::class.java).baseVestingAccount.baseAccount
            this.`is`(Vesting.PermanentLockedAccount::class.java) -> unpack(Vesting.PermanentLockedAccount::class.java).baseVestingAccount.baseAccount
            else -> throw IllegalArgumentException("Account type not handled:$typeUrl")
        }
    }

/**
 * Given an address, get the base account associated with it.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L18).
 *
 * @param bech32Address The bech32 address to fetch.
 * @return [cosmos.auth.v1beta1.Auth.BaseAccount] or throw [IllegalArgumentException] if the account type is not supported.
 */
suspend fun CoroutineAuth.getBaseAccount(bech32Address: String): cosmos.auth.v1beta1.Auth.BaseAccount =
    account(QueryOuterClass.QueryAccountRequest.newBuilder().setAddress(bech32Address).build()).account.run {
        when {
            this.`is`(cosmos.auth.v1beta1.Auth.BaseAccount::class.java) -> unpack(cosmos.auth.v1beta1.Auth.BaseAccount::class.java)
            this.`is`(Vesting.BaseVestingAccount::class.java) -> unpack(Vesting.BaseVestingAccount::class.java).baseAccount
            this.`is`(Vesting.ContinuousVestingAccount::class.java) -> unpack(Vesting.ContinuousVestingAccount::class.java).baseVestingAccount.baseAccount
            this.`is`(Vesting.DelayedVestingAccount::class.java) -> unpack(Vesting.DelayedVestingAccount::class.java).baseVestingAccount.baseAccount
            this.`is`(Vesting.PeriodicVestingAccount::class.java) -> unpack(Vesting.PeriodicVestingAccount::class.java).baseVestingAccount.baseAccount
            this.`is`(Vesting.PermanentLockedAccount::class.java) -> unpack(Vesting.PermanentLockedAccount::class.java).baseVestingAccount.baseAccount
            else -> throw IllegalArgumentException("Account type not handled:$typeUrl")
        }
    }

/**
 * Given an address, get the marker account associated with it.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L26).
 *
 * @param bech32Address The bech32 address to fetch.
 * @return [MarkerAccount] or throw [IllegalArgumentException] if the account type is not supported.
 */
fun BlockingAuth.getMarkerAccount(bech32Address: String): MarkerAccount =
    account(QueryOuterClass.QueryAccountRequest.newBuilder().setAddress(bech32Address).build()).account.run {
        when {
            this.`is`(MarkerAccount::class.java) -> unpack(MarkerAccount::class.java)
            else -> throw IllegalArgumentException("Account type not handled:$typeUrl")
        }
    }

/**
 * Given an address, get the marker account associated with it.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L26).
 *
 * @param bech32Address The bech32 address to fetch.
 * @return [MarkerAccount] or throw [IllegalArgumentException] if the account type is not supported.
 */
suspend fun CoroutineAuth.getMarkerAccount(bech32Address: String): MarkerAccount =
    account(QueryOuterClass.QueryAccountRequest.newBuilder().setAddress(bech32Address).build()).account.run {
        when {
            this.`is`(MarkerAccount::class.java) -> unpack(MarkerAccount::class.java)
            else -> throw IllegalArgumentException("Account type not handled:$typeUrl")
        }
    }
