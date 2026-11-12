package kg.kasapchy.mobile.entities

object DemoData {
    const val sellerId = "azamat"
    const val code = "1234"
    val products = listOf(
        Product(1, "Говяжья мякоть", Category.Beef, 680, "Азамат эт", sellerId, 20,
            "Мякоть без кости. Подойдёт для тушения, плова и домашних котлет."),
        Product(2, "Бараньи рёбрышки", Category.Lamb, 720, "Береке", "bereke", 15,
            "Рёбрышки на кости. Для казана, духовки и приготовления на углях."),
        Product(3, "Курица домашняя", Category.Poultry, 340, "Айыл ферма", "aiyl", 30,
            "Целая охлаждённая курица с фермы. Для бульона и запекания."),
        Product(4, "Говяжий фарш", Category.Beef, 590, "Азамат эт", sellerId, 12,
            "Фарш из говядины для котлет и домашних пельменей."),
    )
    val incomingOrder = Order(1041, products.first(), 2, "aibek@example.com", Delivery.Pickup, "")
}
