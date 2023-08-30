package io.provenance.client.protobuf.extensions

import cosmwasm.wasm.v1.QueryOuterClass
import cosmwasm.wasm.v1.QueryGrpc.QueryBlockingStub as BlockingWasms
import cosmwasm.wasm.v1.QueryGrpcKt.QueryCoroutineStub as CoroutineWasms

/**
 * Query wasm
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L22).
 *
 *  @param request
 *  @return [QueryOuterClass.QuerySmartContractStateResponse].
 */
fun BlockingWasms.queryWasm(request: QueryOuterClass.QuerySmartContractStateRequest): QueryOuterClass.QuerySmartContractStateResponse =
    smartContractState(request)

/**
 * Query wasm
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L22).
 *
 *  @param request
 *  @return [QueryOuterClass.QuerySmartContractStateResponse].
 */
suspend fun CoroutineWasms.queryWasm(request: QueryOuterClass.QuerySmartContractStateRequest): QueryOuterClass.QuerySmartContractStateResponse =
    smartContractState(request)

/**
 * Get contract information.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L25).
 *
 * @param request
 * @return [QueryOuterClass.QueryContractInfoResponse].
 */
fun BlockingWasms.getContractInfo(request: QueryOuterClass.QueryContractInfoRequest): QueryOuterClass.QueryContractInfoResponse =
    contractInfo(request)

/**
 * Get contract information.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L25).
 *
 * @param request
 * @return [QueryOuterClass.QueryContractInfoResponse].
 */
suspend fun CoroutineWasms.getContractInfo(request: QueryOuterClass.QueryContractInfoRequest): QueryOuterClass.QueryContractInfoResponse =
    contractInfo(request)

/**
 * Get contract history.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L28).
 *
 * @param request
 * @return [QueryOuterClass.QueryContractHistoryResponse].
 */
fun BlockingWasms.getContractHistory(request: QueryOuterClass.QueryContractHistoryRequest): QueryOuterClass.QueryContractHistoryResponse =
    contractHistory(request)

/**
 * Get contract history.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L28).
 *
 * @param request
 * @return [QueryOuterClass.QueryContractHistoryResponse].
 */
suspend fun CoroutineWasms.getContractHistory(request: QueryOuterClass.QueryContractHistoryRequest): QueryOuterClass.QueryContractHistoryResponse =
    contractHistory(request)

/**
 * Get contracts by code.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L31).
 *
 * @param request
 * @Return [QueryOuterClass.QueryContractsByCodeResponse].
 */
fun BlockingWasms.getContractsByCode(request: QueryOuterClass.QueryContractsByCodeRequest): QueryOuterClass.QueryContractsByCodeResponse =
    contractsByCode(request)

/**
 * Get contracts by code.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L31).
 *
 * @param request
 * @Return [QueryOuterClass.QueryContractsByCodeResponse].
 */
suspend fun CoroutineWasms.getContractsByCode(request: QueryOuterClass.QueryContractsByCodeRequest): QueryOuterClass.QueryContractsByCodeResponse =
    contractsByCode(request)

/**
 * Get code.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L34).
 *
 * @param request
 * @return [QueryOuterClass.QueryCodesResponse].
 */
fun BlockingWasms.getCode(request: QueryOuterClass.QueryCodeRequest): QueryOuterClass.QueryCodeResponse = code(request)

/**
 * Get code.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L34).
 *
 * @param request
 * @return [QueryOuterClass.QueryCodesResponse].
 */
suspend fun CoroutineWasms.getCode(request: QueryOuterClass.QueryCodeRequest): QueryOuterClass.QueryCodeResponse = code(request)

/**
 * Get codes.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L36).
 *
 * @param request
 * @return [QueryOuterClass.QueryCodesResponse].
 */
fun BlockingWasms.getCodes(request: QueryOuterClass.QueryCodesRequest): QueryOuterClass.QueryCodesResponse = codes(request)

/**
 * Get codes.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L36).
 *
 * @param request
 * @return [QueryOuterClass.QueryCodesResponse].
 */
suspend fun CoroutineWasms.getCodes(request: QueryOuterClass.QueryCodesRequest): QueryOuterClass.QueryCodesResponse = codes(request)
