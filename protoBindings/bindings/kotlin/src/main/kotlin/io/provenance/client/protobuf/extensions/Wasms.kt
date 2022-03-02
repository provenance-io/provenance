package io.provenance.client.protobuf.extensions

import cosmwasm.wasm.v1.QueryOuterClass
import cosmwasm.wasm.v1.QueryGrpc.QueryBlockingStub as Wasms

/**
 * Query wasm
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L22).
 *
 *  @param request
 *  @return [QueryOuterClass.QuerySmartContractStateResponse].
 */
fun Wasms.queryWasm(request: QueryOuterClass.QuerySmartContractStateRequest): QueryOuterClass.QuerySmartContractStateResponse =
    smartContractState(request)

/**
 * Get contract information.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L25).
 *
 * @param request
 * @return [QueryOuterClass.QueryContractInfoResponse].
 */
fun Wasms.getContractInfo(request: QueryOuterClass.QueryContractInfoRequest): QueryOuterClass.QueryContractInfoResponse =
    contractInfo(request)

/**
 * Get contract history.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L28).
 *
 * @param request
 * @return [QueryOuterClass.QueryContractHistoryResponse].
 */
fun Wasms.getContractHistory(request: QueryOuterClass.QueryContractHistoryRequest): QueryOuterClass.QueryContractHistoryResponse =
    contractHistory(request)

/**
 * Get contracts by code.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L31).
 *
 * @param request
 * @Return [QueryOuterClass.QueryContractsByCodeResponse].
 */
fun Wasms.getContractsByCode(request: QueryOuterClass.QueryContractsByCodeRequest): QueryOuterClass.QueryContractsByCodeResponse =
    contractsByCode(request)

/**
 * Get code.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L34).
 *
 * @param request
 * @return [QueryOuterClass.QueryCodesResponse].
 */
fun Wasms.getCode(request: QueryOuterClass.QueryCodeRequest): QueryOuterClass.QueryCodeResponse = code(request)

/**
 * Get codes.
 *
 * See [Wasms](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Wasms.kt#L36).
 *
 * @param request
 * @return [QueryOuterClass.QueryCodesResponse].
 */
fun Wasms.getCodes(request: QueryOuterClass.QueryCodesRequest): QueryOuterClass.QueryCodesResponse = codes(request)
